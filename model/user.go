package model

import (
	"errors"

	"gopkg.in/mgo.v2/bson"
)

// User of the system
type User struct {
	ID          bson.ObjectId `json:"-"             bson:"_id,omitempty"`
	FirebaseUID string        `json:"-"             bson:"firebase-uid"`
	DisplayName string        `json:"display-name"  bson:"display-name"`
	Email       string        `json:"-"             bson:"email"`
	PhotoURL    string        `json:"photo-url"     bson:"photo-url"`
	IsAnon      bool          `json:"is-anon"       bson:"is-anon"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (u User) GetID() string {
	return u.ID.Hex()
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (u *User) SetID(id string) error {

	if id == "" {
		return nil
	}

	if bson.IsObjectIdHex(id) {
		u.ID = bson.ObjectIdHex(id)
		return nil
	}

	return errors.New("<id>" + id + "</id> is not a valid user id")
}
