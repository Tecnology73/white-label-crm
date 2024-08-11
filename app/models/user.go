package models

import (
	"white-label-crm/database"
)

type User struct {
	database.Model `bson:",inline"`

	Name     string `json:"name" bson:"name"`
	Email    string `json:"email" bson:"email"`
	Password string `json:"-" bson:"password"`
}

func (u *User) GetCollectionName() string { return "users" }
