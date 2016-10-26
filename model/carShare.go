package model

import (
	"errors"

	"github.com/manyminds/api2go/jsonapi"
)

// A user of the system
type CarShare struct {
	ID       string   `json:"-"`
	Name     string   `json:"name"`
	Admins   []*User  `json:"-"`
	AdminIDs []string `json:"-"`
	Trips    []*Trip  `json:"-"`
	TripIDs  []string `json:"-"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (cs CarShare) GetID() string {
	return cs.ID
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (cs *CarShare) SetID(id string) error {
	cs.ID = id
	return nil
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
		cs.TripIDs = IDs
		return nil
	}
	if name == "admins" {
		cs.AdminIDs = IDs
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

// AddToManyIDs adds some new trips
func (cs *CarShare) AddToManyIDs(name string, IDs []string) error {
	if name == "trips" {
		cs.TripIDs = append(cs.TripIDs, IDs...)
		return nil
	}
	if name == "adminss" {
		cs.AdminIDs = append(cs.AdminIDs, IDs...)
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

// DeleteToManyIDs removes some sweets from a users because they made him very sick
func (cs *CarShare) DeleteToManyIDs(name string, IDs []string) error {
	if name == "trips" {
		for _, ID := range IDs {
			for pos, oldID := range cs.TripIDs {
				if ID == oldID {
					// match, this ID must be removed
					cs.TripIDs = append(cs.TripIDs[:pos], cs.TripIDs[pos+1:]...)
				}
			}
		}
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
