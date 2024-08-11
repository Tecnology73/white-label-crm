package database

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
	"time"
)

type CollectionModel interface {
	GetCollectionName() string
	SetPrimaryKey(value interface{})
}

type Model struct {
	ID        primitive.ObjectID `json:"_id" bson:"_id"`
	CreatedAt time.Time          `json:"createdAt" bson:"createdAt"`
	UpdatedAt time.Time          `json:"updatedAt" bson:"updatedAt"`
	DeletedAt *time.Time         `json:"deletedAt,omitempty" bson:"deletedAt,omitempty"`
}

func (m *Model) SetPrimaryKey(value interface{}) {
	// TODO: Handle this panic.
	m.ID = value.(primitive.ObjectID)
}
