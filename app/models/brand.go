package models

import (
	"white-label-crm/database"
)

type Brand struct {
	database.Model `bson:",inline"`

	Name string `json:"name" bson:"name"`
	Slug string `json:"slug" bson:"slug"`
}

func (b *Brand) GetCollectionName() string { return "brands" }
