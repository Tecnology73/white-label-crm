package database

import (
	"github.com/gofiber/fiber/v2"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type CollectionModel interface {
	GetCollectionName() string
	SetPrimaryKey(value interface{})
	GetQueryFilter() bson.M

	OnInserted(createdAt time.Time, updatedAt time.Time, updatedBy UserRelation)
	OnUpdated(updatedAt time.Time, updatedBy UserRelation)
}

type UserRelation struct {
	ID   int    `json:"id" bson:"id"`
	Name string `json:"name" bson:"name"`
}

type Model struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	UpdatedBy UserRelation       `json:"updatedBy" bson:"updatedBy"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}

func NewModel(ctx *fiber.Ctx) Model {
	return Model{
		ID:        primitive.NewObjectID(),
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
		UpdatedBy: ctx.Locals("user").(UserRelation),
	}
}

func (m *Model) SetPrimaryKey(value interface{}) {
	// TODO: Handle this panic.
	m.ID = value.(primitive.ObjectID)
}

func (m *Model) GetQueryFilter() bson.M {
	return bson.M{"_id": m.ID}
}

func (m *Model) OnInserted(createdAt time.Time, updatedAt time.Time, updatedBy UserRelation) {
	m.CreatedAt = createdAt
	m.UpdatedAt = updatedAt
	m.UpdatedBy = updatedBy
}

func (m *Model) OnUpdated(updatedAt time.Time, updatedBy UserRelation) {
	m.UpdatedAt = updatedAt
	m.UpdatedBy = updatedBy
}
