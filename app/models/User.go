package models

import "white-label-crm/app/database"

type User struct {
	database.Model `bson:",inline"`

	Name  string `json:"name,omitempty" bson:"name,omitempty"`
	Email string `json:"email,omitempty" bson:"email,omitempty"`
}

func (u *User) GetCollectionName() string { return "users" }
