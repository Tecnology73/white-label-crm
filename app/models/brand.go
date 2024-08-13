package models

import (
	"white-label-crm/database"
)

type Brand struct {
	database.Model `bson:",inline"`

	Name   string `json:"name" bson:"name"`
	Slug   string `json:"slug" bson:"slug"`
	Domain string `json:"domain" bson:"domain"`
}

func (b *Brand) GetCollectionName() string { return "brands" }
