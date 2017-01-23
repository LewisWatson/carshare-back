package memory

import (
	"sort"

	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
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
func NewTripStorage() *TripStorage {
	return &TripStorage{make(map[string]*model.Trip)}
}

// TripStorage in memory trip store
type TripStorage struct {
	trips map[string]*model.Trip
}

// GetAll to satisfy storage.TripStorage interface
func (s TripStorage) GetAll(context api2go.APIContexter) ([]model.Trip, error) {
	result := []model.Trip{}
	for key := range s.trips {
		result = append(result, *s.trips[key])
	}

	sort.Sort(byID(result))
	return result, nil
}

// GetOne to satisfy storage.TripStorage interface
func (s TripStorage) GetOne(id string, context api2go.APIContexter) (model.Trip, error) {
	trip, ok := s.trips[id]
	if !ok {
		return model.Trip{}, storage.ErrNotFound
	}
	return *trip, nil
}

// Insert to satisfy storage.TripStorage interface
func (s *TripStorage) Insert(t model.Trip, context api2go.APIContexter) (string, error) {
	t.ID = bson.NewObjectId()
	s.trips[t.GetID()] = &t
	return t.GetID(), nil
}

// Delete to satisfy storage.TripStorage interface
func (s *TripStorage) Delete(id string, context api2go.APIContexter) error {
	_, exists := s.trips[id]
	if !exists {
		return storage.ErrNotFound
	}
	delete(s.trips, id)

	return nil
}

// Update to satisfy storage.TripStorage interface
func (s *TripStorage) Update(t model.Trip, context api2go.APIContexter) error {
	_, exists := s.trips[t.GetID()]
	if !exists {
		return storage.ErrNotFound
	}
	s.trips[t.GetID()] = &t

	return nil
}

// GetLatest to satisfy storage.TripStorage interface
func (s *TripStorage) GetLatest(carShareID string, context api2go.APIContexter) (model.Trip, error) {

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
