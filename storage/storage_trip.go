package storage

import (
	"errors"
	"fmt"
	"net/http"
	"sort"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// sorting
type byID []model.Trip

func (t byID) Len() int {
	return len(t)
}

func (t byID) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t byID) Less(i, j int) bool {
	return t[i].GetID() < t[j].GetID()
}

// NewTripStorage initializes the storage
func NewTripStorage(db *mgo.Session) *TripStorage {
	return &TripStorage{db.DB("carshare").C("trips"), 1}
}

type TripStorage struct {
	trips   *mgo.Collection
	idCount int
}

// GetAll of the trips
func (s TripStorage) GetAll() ([]model.Trip, error) {
	result := []model.Trip{}
	err := s.trips.Find(nil).All(&result)
	if err != nil {
		errMessage := fmt.Sprintf("Error retrieving trips %s", err)
		return result, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	sort.Sort(byID(result))
	s.setTimezonesToUTC(&result)
	return result, nil
}

// GetOne trip
func (s TripStorage) GetOne(id string) (model.Trip, error) {
	result := model.Trip{}
	err := s.trips.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		errMessage := fmt.Sprintf("Error retrieving trip %s, %s", id, err)
		return result, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	s.setTimezoneToUTC(&result)
	return result, nil
}

// Insert a fresh one
func (s *TripStorage) Insert(t model.Trip) string {
	id := fmt.Sprintf("%d", s.idCount)
	t.ID = id
	s.trips.Insert(&t)
	s.idCount++
	return id
}

// Delete one :(
func (s *TripStorage) Delete(id string) error {
	err := s.trips.Remove(bson.M{"id": id})
	if err != nil {
		errMessage := fmt.Sprintf("Error deleting trip %s, %s", id, err)
		return api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return nil
}

// Update updates an existing trip
func (s *TripStorage) Update(t model.Trip) error {
	err := s.trips.Update(bson.M{"id": t.GetID()}, &t)
	if err != nil {
		errMessage := fmt.Sprintf("Error updating trip %s, %s", t.GetID(), err)
		return api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return nil
}

func (s *TripStorage) GetLatest(carShareID string) (model.Trip, error) {

	trips, err := s.GetAll()
	if err != nil {
		return model.Trip{}, err
	}

	latestTrip := model.Trip{}
	for _, trip := range trips {
		if trip.CarShareID == carShareID {
			if trip.TimeStamp.After(latestTrip.TimeStamp) {
				latestTrip = trip
			}
		}
	}

	s.setTimezoneToUTC(&latestTrip)
	return latestTrip, nil
}

func (s *TripStorage) setTimezonesToUTC(trips *[]model.Trip) {
	for _, trip := range *trips {
		s.setTimezoneToUTC(&trip)
	}
}

// time.Time values get stored in MongoDB as timestamps without timezones
// when they are read they are given a timezone, we want to ensure we stick
// to UTC at all times
func (s *TripStorage) setTimezoneToUTC(trip *model.Trip) {
	trip.TimeStamp = trip.TimeStamp.UTC()
}
