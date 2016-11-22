package in_memory_storage

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
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

// GetAll returns the carShare map (because we need the ID as key too)
func (s CarShareStorage) GetAll() ([]model.CarShare, error) {
	result := []model.CarShare{}
	for key := range s.carShares {
		result = append(result, *s.carShares[key])
	}
	return result, nil
}

// GetOne carShare
func (s CarShareStorage) GetOne(id string) (model.CarShare, error) {
	carShare, ok := s.carShares[id]
	if ok {
		return *carShare, nil
	}
	errMessage := fmt.Sprintf("Car Share for id %s not found", id)
	return model.CarShare{}, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
}

// Insert a carShare
func (s *CarShareStorage) Insert(c model.CarShare) (string, error) {
	c.ID = bson.NewObjectId()
	s.carShares[c.GetID()] = &c
	return c.GetID(), nil
}

// Delete one :(
func (s *CarShareStorage) Delete(id string) error {
	_, exists := s.carShares[id]
	if !exists {
		return fmt.Errorf("Car share with id %s does not exist", id)
	}
	delete(s.carShares, id)

	return nil
}

// Update a carShare
func (s *CarShareStorage) Update(c model.CarShare) error {
	_, exists := s.carShares[c.GetID()]
	if !exists {
		return fmt.Errorf("Car share with id %s does not exist", c.ID)
	}
	s.carShares[c.GetID()] = &c

	return nil
}
