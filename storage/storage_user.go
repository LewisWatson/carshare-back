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

// NewUserStorage initializes the storage
func NewUserStorage(db *mgo.Session) *UserStorage {
	return &UserStorage{db.DB("carshare").C("users"), 1}
}

// UserStorage stores all users
type UserStorage struct {
	users   *mgo.Collection
	idCount int
}

// GetAll of the users
func (s UserStorage) GetAll() ([]model.User, error) {
	result := []model.User{}
	err := s.users.Find(nil).All(&result)
	if err != nil {
		errMessage := fmt.Sprintf("Error retrieving users %s", err)
		return result, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return result, nil
}

// GetOne user
func (s UserStorage) GetOne(id string) (model.User, error) {
	result := model.User{}
	err := s.users.Find(bson.M{"id": id}).One(&result)
	if err != nil {
		errMessage := fmt.Sprintf("Error retrieving user %s, %s", id, err)
		return result, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return result, nil
}

// Insert a user
func (s *UserStorage) Insert(u model.User) string {
	id := fmt.Sprintf("%d", s.idCount)
	u.ID = id
	s.users.Insert(&u)
	s.idCount++
	return id
}

// Delete one :(
func (s *UserStorage) Delete(id string) error {
	err := s.users.Remove(bson.M{"id": id})
	if err != nil {
		errMessage := fmt.Sprintf("Error deleting user %s, %s", id, err)
		return api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return nil
}

// Update a user
func (s *UserStorage) Update(u model.User) error {
	err := s.users.Update(bson.M{"id": u.GetID()}, &u)
	if err != nil {
		errMessage := fmt.Sprintf("Error updating user %s, %s", u.GetID(), err)
		return api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return nil
}
