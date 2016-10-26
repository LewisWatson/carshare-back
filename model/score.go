package model

import (
	"errors"

	"github.com/manyminds/api2go/jsonapi"
)

// A score keeps track of how many miles a user has travelled as a driver and as a passenger
type Score struct {
	ID                string `json:"-"`
	User              *User  `json:"-"`
	UserID            string `json:"-"`
	MetersAsDriver    int    `json:"meters-as-driver"`
	MetersAsPassenger int    `json:"meters-as-passenger"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (s Score) GetID() string {
	return s.ID
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (s *Score) SetID(id string) error {
	s.ID = id
	return nil
}

func (s Score) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type: "users",
			Name: "user",
		},
	}
}

func (s Score) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}

	if s.User != nil {
		result = append(result, jsonapi.ReferenceID{
			ID:   s.User.GetID(),
			Name: "user",
			Type: "Users",
		})
	}

	return result
}

// UnmarshalToOneRelations must be implemented to unmarshal to-one relations
func (s *Score) SetToOneReferenceID(name, ID string) error {
	if name == "user" {
		s.UserID = ID
		return nil
	}

	return errors.New("There is no to-one relationship with the name " + name)
}

func (s Score) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}

	if s.User != nil {
		result = append(result, *s.User)
	}

	return result
}
