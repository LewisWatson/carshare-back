package memory

import (
	"fmt"
	"log"
	"sort"

	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// TripStorage in memory trip store
type TripStorage struct {
	CarShareStorage CarShareStorage
}

// GetAll to satisfy storage.TripStorage interface
func (s *TripStorage) GetAll(carShareID string, context api2go.APIContexter) (map[string]model.Trip, error) {
	carShare, err := s.CarShareStorage.GetOne(carShareID, context)
	if err != nil {
		return nil, err
	}
	return carShare.Trips, nil
}

// GetOne to satisfy storage.TripStorage interface
func (s *TripStorage) GetOne(carShareID string, id string, context api2go.APIContexter) (model.Trip, error) {
	carShare, err := s.CarShareStorage.GetOne(carShareID, context)
	if err != nil {
		return model.Trip{}, err
	}
	trip, ok := carShare.Trips[id]
	if !ok {
		return model.Trip{}, storage.ErrNotFound
	}
	return trip, nil
}

// Insert to satisfy storage.TripStorage interface
func (s *TripStorage) Insert(carShareID string, t model.Trip, context api2go.APIContexter) (string, error) {
	carShare, err := s.CarShareStorage.GetOne(carShareID, context)
	if err != nil {
		return "", err
	}
	t.ID = bson.NewObjectId()
	carShare.Trips[t.GetID()] = t
	return t.GetID(), nil
}

// Delete to satisfy storage.TripStorage interface
func (s TripStorage) Delete(carShareID string, id string, context api2go.APIContexter) error {
	carShare, err := s.CarShareStorage.GetOne(carShareID, context)
	if err != nil {
		return err
	}
	_, exists := carShare.Trips[id]
	if !exists {
		return storage.ErrNotFound
	}
	delete(carShare.Trips, id)
	return nil
}

// Update to satisfy storage.TripStorage interface
func (s *TripStorage) Update(carShareID string, t model.Trip, context api2go.APIContexter) error {
	carShare, err := s.CarShareStorage.GetOne(carShareID, context)
	if err != nil {
		return err
	}
	_, exists := carShare.Trips[t.GetID()]
	if !exists {
		return storage.ErrNotFound
	}
	carShare.Trips[t.GetID()] = t
	return nil
}

// GetLatest to satisfy storage.TripStorage interface
func (s *TripStorage) GetLatest(carShareID string, context api2go.APIContexter) (model.Trip, error) {
	carShare, err := s.CarShareStorage.GetOne(carShareID, context)
	if err != nil {
		log.Printf("Error finding car share %s, %s", carShareID, err)
		return model.Trip{}, err
	}
	if carShare.Trips == nil {
		return model.Trip{}, storage.ErrNotFound
	}
	// sorting keys alphabetically will push the most recent trip to end of the slice
	var keys []string
	for k := range carShare.Trips {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	trip, ok := carShare.Trips[keys[len(keys)-1]]
	if !ok {
		err = fmt.Errorf("Error retrieving latest trip from sorted keys for car share %s", carShareID)
		log.Fatal(err)
		return model.Trip{}, err
	}
	return trip, nil
}
