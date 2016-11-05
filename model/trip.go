package model

import (
	"errors"
	"time"

	"github.com/manyminds/api2go/jsonapi"
)

// A trip is a single instance of a car share
type Trip struct {
	ID         string    `json:"-"`
	Metres     int       `json:"metres"`
	TimeStamp  time.Time `json:"timestamp"`
	CarShare   *CarShare `json:"-"`
	Driver     *User     `json:"-"`
	Passengers []*User   `json:"-"`
	Scores     []*Score  `json:"scores"`
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
			Name: "passengers",
		})
	}

	return result
}

// UnmarshalToOneRelations must be implemented to unmarshal to-one relations
func (t *Trip) SetToOneReferenceID(name, ID string) error {

	if name == "carShare" {
		t.CarShare = &CarShare{ID: ID, Name: "Trip.SetToOneReferenceId temp name"}
		return nil
	}

	if name == "driver" {
		t.Driver = &User{ID: ID, Username: "Trip.SetToOneReferenceID temp username"}
		t.Scores = append(t.Scores, &Score{UserId: t.Driver.GetID(), MetersAsDriver: t.Metres})
		return nil
	}

	return errors.New("There is no to-one relationship with the name " + name)
}

func (t Trip) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}

	/*if t.CarShare != nil {
		result = append(result, *t.CarShare)
	}*/

	if t.Driver != nil {
		result = append(result, *t.Driver)
	}

	for _, passenger := range t.Passengers {
		result = append(result, passenger)
	}

	return result
}

// SetToManyReferenceIDs sets the trips reference IDs and satisfies the jsonapi.UnmarshalToManyRelations interface
func (t *Trip) SetToManyReferenceIDs(name string, IDs []string) error {
	if name == "passengers" {
		t.Passengers = nil
		for _, passengerId := range IDs {
			passenger := &User{ID: passengerId, Username: "Trip.SetToManyReferenceID temp username"}
			t.Passengers = append(t.Passengers, passenger)
			t.Scores = append(t.Scores, &Score{UserId: passenger.GetID(), MetersAsPassenger: t.Metres})
		}
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

// AddToManyIDs adds some new relationships
func (t *Trip) AddToManyIDs(name string, IDs []string) error {
	if name == "passengers" {
		for _, passengerId := range IDs {
			passenger := &User{ID: passengerId}
			t.Passengers = append(t.Passengers, passenger)
			t.Scores = append(t.Scores, &Score{UserId: passenger.GetID(), MetersAsPassenger: t.Metres})
		}
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

// DeleteToManyIDs removes some sweets from a users because they made him very sick
func (t *Trip) DeleteToManyIDs(name string, IDs []string) error {
	if name == "passengers" {
		for _, ID := range IDs {
			for pos, passenger := range t.Passengers {
				if ID == passenger.GetID() {
					// match, this ID must be removed
					t.Passengers = append(t.Passengers[:pos], t.Passengers[pos+1:]...)
				}
			}
			for pos, score := range t.Scores {
				if ID == score.UserId {
					// match, this score must be removed
					t.Scores = append(t.Scores[:pos], t.Scores[pos+1:]...)
				}
			}
		}
	}

	return errors.New("There is no to-many relationship with the name " + name)
}
