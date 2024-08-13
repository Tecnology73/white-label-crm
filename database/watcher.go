package database

import (
	"context"
	"errors"
	"fmt"
	redis2 "github.com/redis/go-redis/v9"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"log"
	"slices"
	"white-label-crm/redis"
)

type Watcher struct {
	client *mongo.Client
	db     *mongo.Database
	stream *mongo.ChangeStream
	cancel context.CancelFunc
}

func NewWatcher(client *mongo.Client) *Watcher {
	watcher := &Watcher{
		client: client,
		db:     GetSystemDb(),
	}

	// Watch for any document changes in `system.brands`
	stream, err := watcher.db.Watch(
		context.Background(),
		mongo.Pipeline{
			{
				{
					"$match", bson.M{
						"ns.db":   "system",
						"ns.coll": "brands",
						"operationType": bson.M{
							"$in": bson.A{"insert", "update", "replace", "delete"},
						},
					},
				},
			},
		},
	)
	if err != nil {
		log.Fatalf("[watcher.NewConnection] %v\n", err)
	}

	// Create a context that will let the goroutine be stopped
	ctx, cancel := context.WithCancel(context.Background())
	watcher.cancel = cancel
	watcher.stream = stream

	go watcher.process(ctx)
	return watcher
}

func (w *Watcher) CloseConnection() {
	if w.stream == nil {
		return
	}

	// Stop processing the change stream
	w.cancel()

	// Close the change stream
	err := w.stream.Close(context.Background())
	if err != nil {
		log.Fatalf("[watcher.CloseConnection] %v\n", err)
	}
}

func (w *Watcher) process(ctx context.Context) {
	for w.stream.Next(ctx) {
		var data changeEvent
		if err := w.stream.Decode(&data); err != nil {
			log.Printf("[Watcher.process] Error: %v\n", err)
			continue
		}

		switch data.OperationType {
		case OperationInsert:
			brand := map[string]string{
				"_id":    data.DocumentKey.ID.Hex(),
				"name":   data.FullDocument["name"].(string),
				"slug":   data.FullDocument["slug"].(string),
				"domain": data.FullDocument["domain"].(string),
			}

			err := w.insertCachedBrand(ctx, data.DocumentKey.ID, brand)
			if err != nil {
				log.Printf("[Watcher.process] Insert | %v\n", err)
			}
		case OperationUpdate:
			fallthrough
		case OperationReplace:
			err := w.updateCachedBrand(ctx, data)
			if err != nil {
				log.Printf("[Watcher.process] Update | %v\n", err)
			}
		case OperationDelete:
			err := w.deleteCachedBrand(ctx, data)
			if err != nil {
				log.Printf("[Watcher.process] Delete | %v\n", err)
			}
		default:
			log.Printf("[Watcher.process] Unhandled event: %v\n", data.OperationType)
		}
	}
}

func wasSoftDeleted(data changeEvent) bool {
	_, ok := data.UpdateDescription.UpdatedFields["deletedAt"].(primitive.DateTime)
	return ok
}

func wasRestored(data changeEvent) bool {
	return slices.Contains(data.UpdateDescription.RemovedFields, "deletedAt")
}

func getChanges(data changeEvent) map[string]string {
	changes := map[string]string{}

	if name, hasChanged := data.UpdateDescription.UpdatedFields["name"].(string); hasChanged {
		changes["name"] = name
	}

	return changes
}

func (w *Watcher) insertCachedBrand(ctx context.Context, id primitive.ObjectID, brand map[string]string) error {
	_, err := redis.Client.TxPipelined(
		ctx,
		func(pipe redis2.Pipeliner) error {
			key := fmt.Sprintf("brands:%v", brand["domain"])
			for k, v := range brand {
				pipe.HSet(ctx, key, k, v)
			}

			// For reverse lookup
			pipe.Set(
				ctx,
				fmt.Sprintf("brands:$id:%s", id.Hex()),
				brand["domain"],
				0,
			)

			return nil
		},
	)

	return err
}

func (w *Watcher) updateCachedBrand(ctx context.Context, data changeEvent) error {
	if wasSoftDeleted(data) {
		return w.deleteCachedBrand(ctx, data)
	}

	if wasRestored(data) {
		return w.restoreCachedBrand(ctx, data)
	}

	// If the domain changed, the keys need to be updated.
	// Just delete the old cached data & recreate it.
	if _, hasChanged := data.UpdateDescription.UpdatedFields["domain"].(string); hasChanged {
		if err := w.deleteCachedBrand(ctx, data); err != nil {
			return err
		}

		return w.restoreCachedBrand(ctx, data)
	}

	// Get a list of changes that we care about.
	changes := getChanges(data)
	if len(changes) == 0 {
		return nil
	}

	domain, err := redis.Client.Get(ctx, fmt.Sprintf("brands:$id:%s", data.DocumentKey.ID.Hex())).Result()
	if err != nil {
		if errors.Is(err, redis2.Nil) {
			return nil
		}

		return err
	}

	_, err = redis.Client.TxPipelined(
		ctx,
		func(pipe redis2.Pipeliner) error {
			key := fmt.Sprintf("brands:%s", domain)
			for k, v := range changes {
				pipe.HSet(ctx, key, k, v)
			}

			return nil
		},
	)

	return err
}

func (w *Watcher) restoreCachedBrand(ctx context.Context, data changeEvent) error {
	var brand bson.M
	err := GetSystemDb().Collection("brands").
		FindOne(ctx, bson.M{"_id": data.DocumentKey.ID}).
		Decode(&brand)
	if err != nil {
		if errors.Is(err, mongo.ErrNoDocuments) {
			return nil
		}

		return err
	}

	return w.insertCachedBrand(
		ctx,
		brand["_id"].(primitive.ObjectID),
		map[string]string{
			"_id":    brand["_id"].(primitive.ObjectID).Hex(),
			"name":   brand["name"].(string),
			"slug":   brand["slug"].(string),
			"domain": brand["domain"].(string),
		},
	)
}

func (w *Watcher) deleteCachedBrand(ctx context.Context, data changeEvent) error {
	// Lookup (and delete) the domain from _id
	domain, err := redis.Client.GetDel(ctx, fmt.Sprintf("brands:$id:%s", data.DocumentKey.ID.Hex())).Result()
	if err != nil {
		if errors.Is(err, redis2.Nil) {
			return nil
		}

		return err
	}

	// Delete brand info
	return redis.Client.Del(ctx, fmt.Sprintf("brands:%s", domain)).Err()
}

type OperationType string

const (
	OperationInsert  OperationType = "insert"
	OperationUpdate  OperationType = "update"
	OperationReplace OperationType = "replace"
	OperationDelete  OperationType = "delete"
)

type changeEvent struct {
	ID                bson.D                 `bson:"_id"`
	OperationType     OperationType          `bson:"operationType"`
	DocumentKey       documentKey            `bson:"documentKey"`
	FullDocument      map[string]interface{} `bson:"fullDocument"`
	UpdateDescription updateDescription      `bson:"updateDescription"`
	Namespace         namespace              `bson:"ns"`
}

type documentKey struct {
	ID primitive.ObjectID `bson:"_id"`
}

type updateDescription struct {
	UpdatedFields map[string]interface{} `bson:"updatedFields"`
	RemovedFields []string               `bson:"removedFields"`
	//TruncatedArrays    []truncatedField       `bson:"truncatedFields"`
	//DisambiguatedPaths map[string]interface{} `bson:"disambiguatedPaths"`
}

/*type truncatedField struct {
	Field   string `bson:"field"`
	NewSize uint32 `bson:"newSize"`
}*/

type namespace struct {
	Database   string `bson:"db"`
	Collection string `bson:"coll"`
}
