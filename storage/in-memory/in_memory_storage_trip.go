package in_memory_storage

import (
	"fmt"
	"sort"

	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
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
func NewTripStorage() *TripStorage {
	return &TripStorage{make(map[string]*model.Trip)}
}

type TripStorage struct {
	trips map[string]*model.Trip
}

// GetAll of the trips
func (s TripStorage) GetAll() ([]model.Trip, error) {
	result := []model.Trip{}
	for key := range s.trips {
		result = append(result, *s.trips[key])
	}

	sort.Sort(byID(result))
	return result, nil
}

// GetOne trip
func (s TripStorage) GetOne(id string) (model.Trip, error) {
	trip, ok := s.trips[id]
	if ok {
		return *trip, nil
	}

	return model.Trip{}, fmt.Errorf("Trip for id %s not found", id)
}

// Insert a fresh one
func (s *TripStorage) Insert(t model.Trip) (string, error) {
	t.ID = bson.NewObjectId()
	s.trips[t.GetID()] = &t
	return t.GetID(), nil
}

// Delete one :(
func (s *TripStorage) Delete(id string) error {
	_, exists := s.trips[id]
	if !exists {
		return fmt.Errorf("Trip with id %s does not exist", id)
	}
	delete(s.trips, id)

	return nil
}

// Update updates an existing trip
func (s *TripStorage) Update(t model.Trip) error {
	_, exists := s.trips[t.GetID()]
	if !exists {
		return fmt.Errorf("Trip with id %s does not exist", t.ID)
	}
	s.trips[t.GetID()] = &t

	return nil
}

func (s *TripStorage) GetLatest(carShareID string) (model.Trip, error) {

	latestTrip := model.Trip{}

	for _, trip := range s.trips {
		if trip.CarShareID == carShareID {
			if trip.TimeStamp.After(latestTrip.TimeStamp) {
				latestTrip = *trip
			}
		}
	}

	return latestTrip, nil
}
