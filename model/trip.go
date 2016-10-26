package model

import (
	"errors"
	"time"

	"github.com/manyminds/api2go/jsonapi"
)

// A trip is a single instance of a car share
type Trip struct {
	ID           string    `json:"-"`
	Metres       int       `json:"metres"`
	TimeStamp    time.Time `json:"timestamp"`
	CarShare     *CarShare `json:"-"`
	CarShareID   string    `json:"-"`
	Driver       *User     `json:"-"`
	DriverID     string    `json:"-"`
	Passengers   []*User   `json:"-"`
	PassengerIDs []string  `json:"-"`
	Scores       []*Score  `json:"-"`
	ScoreIDs     []string  `json:"-"`
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
			Type: "carShares",
			Name: "carShare",
		},
		{
			Type: "users",
			Name: "driver",
		},
		{
			Type: "users",
			Name: "passengers",
		},
		{
			Type: "scores",
			Name: "scores",
		},
	}
}

func (t Trip) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}

	if t.CarShare != nil {
		result = append(result, jsonapi.ReferenceID{
			ID:   t.CarShare.GetID(),
			Name: "carShare",
			Type: "carShares",
		})
	}

	if t.Driver != nil {
		result = append(result, jsonapi.ReferenceID{
			ID:   t.Driver.GetID(),
			Name: "driver",
			Type: "users",
		})
	}

	for _, passenger := range t.Passengers {
		result = append(result, jsonapi.ReferenceID{
			ID:   passenger.GetID(),
			Type: "users",
			Name: "passenger",
		})
	}

	for _, score := range t.Scores {
		result = append(result, jsonapi.ReferenceID{
			ID:   score.GetID(),
			Type: "scores",
			Name: "scores",
		})
	}

	return result
}

// UnmarshalToOneRelations must be implemented to unmarshal to-one relations
func (t *Trip) SetToOneReferenceID(name, ID string) error {

	if name == "carShare" {
		t.CarShareID = ID
		return nil
	}

	if name == "driver" {
		t.DriverID = ID
		return nil
	}

	return errors.New("There is no to-one relationship with the name " + name)
}

func (t Trip) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}

	if t.CarShare != nil {
		result = append(result, *t.CarShare)
	}

	if t.Driver != nil {
		result = append(result, *t.Driver)
	}

	for key := range t.Passengers {
		result = append(result, t.Passengers[key])
	}

	for key := range t.Scores {
		result = append(result, t.Scores[key])
	}

	return result
}
