package models

import (
	"white-label-crm/app/database"
)

type Brand struct {
	database.Model `bson:",inline"`
}

func (b *Brand) GetCollectionName() string { return "brands" }
