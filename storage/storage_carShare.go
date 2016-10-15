package storage

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// NewCarShareStorage initializes the storage
func NewCarShareStorage() *CarShareStorage {
	return &CarShareStorage{make(map[string]*model.CarShare), 1}
}

// CarShareStorage stores all car shares
type CarShareStorage struct {
	carShares map[string]*model.CarShare
	idCount   int
}

// GetAll returns the carShare map (because we need the ID as key too)
func (s CarShareStorage) GetAll() map[string]*model.CarShare {
	return s.carShares
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
func (s *CarShareStorage) Insert(c model.CarShare) string {
	id := fmt.Sprintf("%d", s.idCount)
	c.ID = id
	s.carShares[id] = &c
	s.idCount++
	return id
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
	_, exists := s.carShares[c.ID]
	if !exists {
		return fmt.Errorf("Car share with id %s does not exist", c.ID)
	}
	s.carShares[c.ID] = &c

	return nil
}
