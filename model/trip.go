package model

import (
	"errors"
	"sort"
	"time"

	"gopkg.in/mgo.v2/bson"

	"github.com/manyminds/api2go/jsonapi"
)

// A trip is a single instance of a car share
type Trip struct {
	ID           bson.ObjectId    `json:"-"         bson:"_id,omitempty"`
	Metres       int              `json:"metres"    bson:"metres"`
	TimeStamp    time.Time        `json:"timestamp" bson:"timestamp"`
	CarShare     *CarShare        `json:"-"         bson:"-"`
	CarShareID   string           `json:"-"         bson:"car-share"`
	Driver       *User            `json:"-"         bson:"-"`
	DriverID     string           `json:"-"         bson:"driver"`
	Passengers   []*User          `json:"-"         bson:"-"`
	PassengerIDs []string         `json:"-"         bson:"passengers"`
	Scores       map[string]Score `json:"scores"    bson:"scores"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (t Trip) GetID() string {
	return t.ID.Hex()
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (t *Trip) SetID(id string) error {

	// for some reason SetID gets called with null ("") when run in TravisCI
	// this doesn't seem to happen during local builds.
	if id == "" {
		return nil
	}

	if bson.IsObjectIdHex(id) {
		t.ID = bson.ObjectIdHex(id)
		return nil
	}

	return errors.New("<id>" + id + "</id> is not a valid trip id")
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

	if t.CarShareID != "" {
		result = append(result, jsonapi.ReferenceID{
			ID:   t.CarShareID,
			Name: "carShare",
			Type: "carShares",
		})
	}

	if t.DriverID != "" {
		result = append(result, jsonapi.ReferenceID{
			ID:   t.DriverID,
			Name: "driver",
			Type: "users",
		})
	}

	for _, passengerID := range t.PassengerIDs {
		result = append(result, jsonapi.ReferenceID{
			ID:   passengerID,
			Type: "users",
			Name: "passengers",
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
		t.PassengerIDs = nil
		for _, passengerID := range IDs {
			t.PassengerIDs = append(t.PassengerIDs, passengerID)
		}
		sort.Strings(t.PassengerIDs)
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

// AddToManyIDs adds some new relationships
func (t *Trip) AddToManyIDs(name string, IDs []string) error {
	if name == "passengers" {
		for _, passengerID := range IDs {
			t.PassengerIDs = append(t.PassengerIDs, passengerID)
		}
		sort.Strings(t.PassengerIDs)
		return nil
	}
	return errors.New("There is no to-many relationship with the name " + name)
}

// DeleteToManyIDs removes some sweets from a users because they made him very sick
func (t *Trip) DeleteToManyIDs(name string, IDs []string) error {
	if name == "passengers" {
		for _, ID := range IDs {
			for pos, passengerID := range t.PassengerIDs {
				if ID == passengerID {
					// match, this ID must be removed
					t.PassengerIDs = append(t.PassengerIDs[:pos], t.PassengerIDs[pos+1:]...)
				}
			}
		}
		return nil
	}

	return errors.New("There is no to-many relationship with the name " + name)
}

func (t *Trip) CalculateScores(scoresFromLastTrip map[string]Score) error {

	if scoresFromLastTrip != nil {
		t.Scores = scoresFromLastTrip
	}

	if t.DriverID != "" {
		driverScore, ok := t.Scores[t.DriverID]
		if ok {
			driverScore.MetresAsDriver += t.Metres
		} else {
			driverScore = Score{MetresAsDriver: t.Metres, MetresAsPassenger: 0}
		}
		t.Scores[t.DriverID] = driverScore
	}

	for _, passengerID := range t.PassengerIDs {
		passengerScore, ok := t.Scores[passengerID]
		if ok {
			passengerScore.MetresAsPassenger += t.Metres
		} else {
			passengerScore = Score{MetresAsDriver: 0, MetresAsPassenger: t.Metres}
		}
		t.Scores[passengerID] = passengerScore
	}

	return nil
}
