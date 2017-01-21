package memory

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// NewCarShareStorage initializes the storage
func NewCarShareStorage() *CarShareStorage {
	return &CarShareStorage{make(map[string]*model.CarShare)}
}

// CarShareStorage stores all car shares
type CarShareStorage struct {
	carShares map[string]*model.CarShare
}

// GetAll to satisfy storage.CarShareStoreage interface
func (s CarShareStorage) GetAll(context api2go.APIContexter) ([]model.CarShare, error) {
	result := []model.CarShare{}
	for key := range s.carShares {
		result = append(result, *s.carShares[key])
	}
	return result, nil
}

// GetOne to satisfy storage.CarShareStoreage interface
func (s CarShareStorage) GetOne(id string, context api2go.APIContexter) (model.CarShare, error) {
	carShare, ok := s.carShares[id]
	if !ok {
		return model.CarShare{}, storage.ErrNotFound
	}
	return *carShare, nil
}

// Insert to satisfy storage.CarShareStoreage interface
func (s *CarShareStorage) Insert(c model.CarShare, context api2go.APIContexter) (string, error) {
	c.ID = bson.NewObjectId()
	s.carShares[c.GetID()] = &c
	return c.GetID(), nil
}

// Delete to satisfy storage.CarShareStoreage interface
func (s *CarShareStorage) Delete(id string, context api2go.APIContexter) error {
	_, exists := s.carShares[id]
	if !exists {
		return storage.ErrNotFound
	}
	delete(s.carShares, id)

	return nil
}

// Update to satisfy storage.CarShareStoreage interface
func (s *CarShareStorage) Update(c model.CarShare, context api2go.APIContexter) error {
	_, exists := s.carShares[c.GetID()]
	if !exists {
		return storage.ErrNotFound
	}
	s.carShares[c.GetID()] = &c

	return nil
}
