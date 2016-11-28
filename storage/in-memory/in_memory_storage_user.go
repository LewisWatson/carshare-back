package in_memory_storage

import (
	"errors"
	"fmt"
	"net/http"

	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// NewUserStorage initializes the storage
func NewUserStorage() *UserStorage {
	return &UserStorage{make(map[string]*model.User), 1}
}

// UserStorage stores all users
type UserStorage struct {
	users   map[string]*model.User
	idCount int
}

// GetAll of the users
func (s UserStorage) GetAll(context api2go.APIContexter) ([]model.User, error) {
	result := []model.User{}
	for key := range s.users {
		result = append(result, *s.users[key])
	}
	return result, nil
}

// GetOne user
func (s UserStorage) GetOne(id string, context api2go.APIContexter) (model.User, error) {
	user, ok := s.users[id]
	if ok {
		return *user, nil
	}
	errMessage := fmt.Sprintf("User for id %s not found", id)
	return model.User{}, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
}

// Insert a user
func (s *UserStorage) Insert(u model.User, context api2go.APIContexter) (string, error) {
	u.ID = bson.NewObjectId()
	s.users[u.GetID()] = &u
	s.idCount++
	return u.GetID(), nil
}

// Delete one :(
func (s *UserStorage) Delete(id string, context api2go.APIContexter) error {
	_, exists := s.users[id]
	if !exists {
		return fmt.Errorf("User with id %s does not exist", id)
	}
	delete(s.users, id)

	return nil
}

// Update a user
func (s *UserStorage) Update(u model.User, context api2go.APIContexter) error {
	_, exists := s.users[u.GetID()]
	if !exists {
		return fmt.Errorf("User with id %s does not exist", u.ID)
	}
	s.users[u.GetID()] = &u
	return nil
}
