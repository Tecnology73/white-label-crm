package database

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var (
	client *mongo.Client
)

func NewConnection(opts *options.ClientOptions) *mongo.Client {
	var err error
	client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
	}

	return client
}

func CloseConnection() {
	if err := client.Disconnect(context.Background()); err != nil {
		log.Fatalf("[database.CloseConnection] %v\n", err)
	}
	client = nil
}

func GetSystemDb() *mongo.Database {
	return client.Database("system")
}

func GetBrandDb(ctx *fiber.Ctx) *mongo.Database {
	dbName, ok := ctx.Locals("dbName").(string)
	if !ok {
		return nil
	}

	return client.Database(dbName)
}

func Find[T any, R interface {
	*T
	CollectionModel
}](
	db *mongo.Database,
	ctx context.Context,
	filter bson.M,
	opts ...*options.FindOptions,
) ([]R, error) {
	var m R // Temporary
	/*log.Printf(
		"[database.Find] Collection(%v) | Filter(%v)\n",
		m.GetCollectionName(),
		filter,
	)*/

	cursor, err := db.Collection(m.GetCollectionName()).Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	var records []R
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		// Ensure an array is always returned for consistency.
		return []R{}, nil
	}

	return records, nil
}

func FindOne[T any, R interface {
	*T
	CollectionModel
}](
	db *mongo.Database,
	ctx context.Context,
	filter bson.M,
	opts ...*options.FindOneOptions,
) (R, error) {
	var record R
	/*log.Printf(
		"[database.FindOne] Collection(%v) | Filter(%v)\n",
		record.GetCollectionName(),
		filter,
	)*/

	err := db.Collection(record.GetCollectionName()).
		FindOne(ctx, filter, opts...).
		Decode(&record)

	return record, err
}

func InsertOne[T CollectionModel](
	db *mongo.Database,
	ctx context.Context,
	doc interface{},
	opts ...*options.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	var m T // temporary
	/*log.Printf(
		"[database.InsertOne] Collection(%v) | Record(%v)\n",
		m.GetCollectionName(),
		doc,
	)*/

	return db.Collection(m.GetCollectionName()).InsertOne(ctx, doc, opts...)
}

func InsertMany[T CollectionModel](
	db *mongo.Database,
	ctx context.Context,
	docs []interface{},
	opts ...*options.InsertManyOptions,
) (*mongo.InsertManyResult, error) {
	var m T // temporary
	/*log.Printf(
		"[database.InsertMany] Collection(%v) | Docs(%v)\n",
		m.GetCollectionName(),
		docs,
	)*/

	return db.Collection(m.GetCollectionName()).InsertMany(ctx, docs, opts...)
}

func UpdateOne[T CollectionModel](
	db *mongo.Database,
	ctx context.Context,
	filter bson.M,
	update bson.M,
	opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	var m T // temporary
	/*log.Printf(
		"[database.UpdateOne] Collection(%v) | Filter(%v) | Update(%v)\n",
		m.GetCollectionName(),
		filter,
		update,
	)*/

	return db.Collection(m.GetCollectionName()).UpdateOne(ctx, filter, update, opts...)
}

func UpdateMany[T CollectionModel](
	db *mongo.Database,
	ctx context.Context,
	filter bson.M,
	update bson.M,
	opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	var m T // temporary
	/*log.Printf(
		"[database.UpdateMany] Collection(%v) | Filter(%v) | Update(%v)\n",
		m.GetCollectionName(),
		filter,
		update,
	)*/

	return db.Collection(m.GetCollectionName()).UpdateMany(ctx, filter, update, opts...)
}
