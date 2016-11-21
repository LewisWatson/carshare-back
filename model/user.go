package model

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

// A user of the system
type User struct {
	ID       bson.ObjectId `json:"-" bson:"_id,omitempty"`
	Username string        `json:"user-name" bson:"user-name"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (u User) GetID() string {
	return u.ID.Hex()
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (u *User) SetID(id string) error {

	if bson.IsObjectIdHex(id) {
		u.ID = bson.ObjectIdHex(id)
		return nil
	}

	return errors.New(id + " is not a valid id")
}
