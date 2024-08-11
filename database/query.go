package database

import (
	"context"
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"time"
)

type Query struct {
	set         *bson.M
	unset       *bson.M
	currentDate *bson.M

	ctx *fiber.Ctx

	CreatedAt time.Time
	UpdatedAt time.Time
	UpdatedBy UserRelation
}

func NewQuery(ctx *fiber.Ctx) *Query {
	return &Query{
		ctx: ctx,
	}
}

func (q *Query) Set(key string, value interface{}) *Query {
	if q.set == nil {
		q.set = &bson.M{}
	}

	(*q.set)[key] = value

	return q
}

func (q *Query) Unset(key ...string) *Query {
	if q.unset == nil {
		q.unset = &bson.M{}
	}

	for _, k := range key {
		(*q.unset)[k] = true
	}

	return q
}

func (q *Query) CurrentDate(key string, value ...interface{}) *Query {
	if q.currentDate == nil {
		q.currentDate = &bson.M{}
	}

	if len(value) > 0 {
		(*q.currentDate)[key] = value[0]
	} else {
		(*q.currentDate)[key] = true
	}

	return q
}

func (q *Query) Encode() bson.M {
	out := bson.M{}
	if q.set != nil {
		out["$set"] = *q.set
	}

	if q.unset != nil {
		out["$unset"] = *q.unset
	}

	if q.currentDate != nil {
		out["$currentDate"] = *q.currentDate
	}

	return out
}

func (q *Query) EncodeInsert() (*mongo.Database, bson.M) {
	db, out := q.EncodeUpdate()

	q.CreatedAt = time.Now()
	out["$set"].(bson.M)["createdAt"] = q.CreatedAt

	// There are no operators ($set, $currentDate, etc.) for insert queries.
	return db, out["$set"].(bson.M)
}

func (q *Query) EncodeUpdate() (*mongo.Database, bson.M) {
	out := q.Encode()
	set, ok := out["$set"].(bson.M)
	if !ok {
		set = bson.M{}
		out["$set"] = set
	}

	q.UpdatedAt = time.Now()
	set["updatedAt"] = q.UpdatedAt

	q.UpdatedBy = q.ctx.Locals("user").(UserRelation)
	set["updatedBy"] = q.UpdatedBy

	// For when we do an upsert.
	setOnInsert, ok := out["$setOnInsert"].(bson.M)
	if !ok {
		setOnInsert = bson.M{}
		out["$setOnInsert"] = setOnInsert
	}

	q.CreatedAt = time.Now()
	setOnInsert["createdAt"] = q.CreatedAt

	return GetBrandDb(q.ctx), out
}

func (q *Query) InsertOne(
	ctx context.Context,
	record CollectionModel,
	opts ...*options.InsertOneOptions,
) (*mongo.InsertOneResult, error) {
	db, doc := q.EncodeInsert()
	result, err := db.Collection(record.GetCollectionName()).InsertOne(ctx, doc, opts...)
	if err != nil {
		return result, err
	}

	record.SetPrimaryKey(result.InsertedID)
	record.OnInserted(q.CreatedAt, q.UpdatedAt, q.UpdatedBy)

	return result, nil
}

func (q *Query) UpdateOne(
	ctx context.Context,
	record CollectionModel,
	opts ...*options.UpdateOptions,
) (*mongo.UpdateResult, error) {
	db, update := q.EncodeUpdate()
	result, err := db.Collection(record.GetCollectionName()).UpdateOne(ctx, record.GetQueryFilter(), update, opts...)
	if err != nil {
		return result, err
	}

	if result.UpsertedCount == 1 {
		record.SetPrimaryKey(result.UpsertedID)
	}

	if result.ModifiedCount == 1 {
		record.OnUpdated(q.UpdatedAt, q.UpdatedBy)
	}

	return result, nil
}
