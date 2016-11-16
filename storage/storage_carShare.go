package storage

import (
	"errors"
	"fmt"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// NewCarShareStorage initializes the storage
func NewCarShareStorage(db *mgo.Session) *CarShareStorage {
	return &CarShareStorage{db.DB("carshare").C("carShares"), 1}
}

// CarShareStorage stores all car shares
type CarShareStorage struct {
	carShares *mgo.Collection
	idCount   int
}

func (s CarShareStorage) GetAll() ([]model.CarShare, error) {
	result := []model.CarShare{}
	err := s.carShares.Find(nil).All(&result)
	if err != nil {
		errMessage := fmt.Sprintf("Error retrieving carShares %s", err)
		return result, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return result, nil
}

// GetOne carShare
func (s CarShareStorage) GetOne(id string) (model.CarShare, error) {
	result := model.CarShare{}
	err := s.carShares.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		errMessage := fmt.Sprintf("Error retrieving carShare %s, %s", id, err)
		return result, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return result, nil
}

// Insert a carShare
func (s *CarShareStorage) Insert(c model.CarShare) string {
	id := fmt.Sprintf("%d", s.idCount)
	c.ID = id
	s.carShares.Insert(&c)
	s.idCount++
	return id
}

// Delete one :(
func (s *CarShareStorage) Delete(id string) error {
	err := s.carShares.Remove(bson.M{"id": id})
	if err != nil {
		errMessage := fmt.Sprintf("Error deleting carShare %s, %s", id, err)
		return api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return nil
}

// Update a carShare
func (s *CarShareStorage) Update(c model.CarShare) error {
	err := s.carShares.Update(bson.M{"id": c.GetID()}, &c)
	if err != nil {
		errMessage := fmt.Sprintf("Error updating carShare %s, %s", c.GetID(), err)
		return api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return nil
}
