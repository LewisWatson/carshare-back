package model

import (
	"errors"
	"fmt"
	"sort"

	"gopkg.in/mgo.v2/bson"

	"github.com/manyminds/api2go/jsonapi"
)

// CarShare an individual group of users who make up a car share
type CarShare struct {
	ID       bson.ObjectId `json:"-"    bson:"_id,omitempty"`
	Name     string        `json:"name" bson:"name"`
	Admins   []*User       `json:"-"    bson:"-"`
	AdminIDs []string      `json:"-"    bson:"admins"`
	Trips    []Trip        `json:"-"    bson:"-"`
	TripIDs  []string      `json:"-"    bson:"trips"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (cs CarShare) GetID() string {
	return cs.ID.Hex()
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (cs *CarShare) SetID(id string) error {

	// for some reason SetID gets called with null ("") when run in TravisCI
	// this doesn't seem to happen during local builds.
	if id == "" {
		return nil
	}

	if bson.IsObjectIdHex(id) {
		cs.ID = bson.ObjectIdHex(id)
		return nil
	}

	return errors.New("<id>" + id + "</id> is not a valid car share id")
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (cs CarShare) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type: "trips",
			Name: "trips",
		},
		{
			Type: "users",
			Name: "admins",
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (cs CarShare) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}

	for _, trip := range cs.Trips {
		result = append(result, jsonapi.ReferenceID{
			ID:   trip.GetID(),
			Type: "trips",
			Name: "trips",
		})
	}

	for _, admin := range cs.Admins {
		result = append(result, jsonapi.ReferenceID{
			ID:   admin.GetID(),
			Type: "users",
			Name: "admins",
		})
	}

	return result
}

// GetReferencedStructs to satisfy the jsonapi.MarhsalIncludedRelations interface
func (cs CarShare) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}
	for key := range cs.Trips {
		result = append(result, cs.Trips[key])
	}
	for key := range cs.Admins {
		result = append(result, cs.Admins[key])
	}
	return result
}

// SetToManyReferenceIDs sets the trips reference IDs and satisfies the jsonapi.UnmarshalToManyRelations interface
func (cs *CarShare) SetToManyReferenceIDs(name string, IDs []string) error {
	if name == "trips" {
		return fmt.Errorf("unsupported operation")
	}
	if name == "admins" {
		cs.AdminIDs = IDs
		sort.Strings(cs.AdminIDs)
		return nil
	}
	return fmt.Errorf("There is no to-many relationship with the name " + name)
}

// AddToManyIDs adds some new trips
func (cs *CarShare) AddToManyIDs(name string, IDs []string) error {
	if name == "trips" {
		return fmt.Errorf("unsupported operation")
	}
	if name == "admins" {
		cs.AdminIDs = append(cs.AdminIDs, IDs...)
		sort.Strings(cs.AdminIDs)
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

// DeleteToManyIDs removes some relationships from car shrae
func (cs *CarShare) DeleteToManyIDs(name string, IDs []string) error {
	if name == "trips" {
		return fmt.Errorf("unsupported operation")
	}
	if name == "admins" {
		for _, ID := range IDs {
			for pos, oldID := range cs.AdminIDs {
				if ID == oldID {
					// match, this ID must be removed
					cs.TripIDs = append(cs.AdminIDs[:pos], cs.AdminIDs[pos+1:]...)
				}
			}
		}
	}

	return errors.New("There is no to-many relationship with the name " + name)
}
