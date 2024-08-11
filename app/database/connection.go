package database

import (
	"context"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"log"
)

var (
	client   *mongo.Client
	database *mongo.Database
)

func NewConnection(opts *options.ClientOptions) {
	var err error
	client, err = mongo.Connect(context.TODO(), opts)
	if err != nil {
		log.Fatal(err)
	}

	database = client.Database("alpha") // TODO: Change based on brand from request
}

func Find[T CollectionModel](ctx context.Context, filter bson.M, opts ...*options.FindOptions) ([]T, error) {
	var m T // Temporary
	cursor, err := database.Collection(m.GetCollectionName()).Find(ctx, filter, opts...)
	if err != nil {
		return nil, err
	}

	var records []T
	if err = cursor.All(ctx, &records); err != nil {
		return nil, err
	}

	if len(records) == 0 {
		// Ensure an array is always returned for consistency.
		return []T{}, nil
	}

	return records, nil
}

func FindOne[T CollectionModel](ctx context.Context, filter bson.M, opts ...*options.FindOneOptions) (*T, error) {
	var record T
	err := database.Collection(record.GetCollectionName()).
		FindOne(ctx, filter, opts...).
		Decode(&record)
	if err != nil {
		return nil, err
	}

	return &record, nil
}

func Insert[T any, R interface {
	*T
	CollectionModel
}](ctx context.Context, record R, opts ...*options.InsertOneOptions) error {
	result, err := database.Collection(record.GetCollectionName()).InsertOne(ctx, record, opts...)
	if err != nil {
		return err
	}

	record.SetPrimaryKey(result.InsertedID)
	return nil
}

func InsertMany[T any, R interface {
	*T
	CollectionModel
}](ctx context.Context, records []R, opts ...*options.InsertManyOptions) error {
	if len(records) == 0 {
		return nil
	}

	docs := make([]interface{}, len(records))
	for i, record := range records {
		docs[i] = record
	}

	result, err := database.Collection(records[0].GetCollectionName()).InsertMany(ctx, docs, opts...)
	if err != nil {
		return err
	}

	for index, id := range result.InsertedIDs {
		records[index].SetPrimaryKey(id)
	}

	return nil
}
