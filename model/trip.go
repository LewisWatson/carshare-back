package model

import (
	"errors"

	"github.com/manyminds/api2go/jsonapi"
)

// A trip is a single instance of a car share
type Trip struct {
	ID     string `json:"-"`
	User   *User  `json:"-"`
	UserID string `json:"-"`
	// TimeStamp         time.Time `json:"timestamp"`
	MetersAsDriver    int `json:"meters-as-driver"`
	MetersAsPassenger int `json:"meters-as-passenger"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (t Trip) GetID() string {
	return t.ID
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (t *Trip) SetID(id string) error {
	t.ID = id
	return nil
}

func (t Trip) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type: "users",
			Name: "users",
		},
	}
}

func (t Trip) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}

	if t.User != nil {
		result = append(result, jsonapi.ReferenceID{
			ID:   t.User.GetID(),
			Name: "user",
			Type: "Users",
		})
	}

	return result
}

// UnmarshalToOneRelations must be implemented to unmarshal to-one relations
func (t *Trip) SetToOneReferenceID(name, ID string) error {
	if name == "user" {
		t.UserID = ID
		return nil
	}

	return errors.New("There is no to-one relationship with the name " + name)
}

func (t Trip) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}

	if t.User != nil {
		result = append(result, *t.User)
	}

	return result
}
