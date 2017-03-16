package memory

import (
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
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
	if !ok {
		return model.User{}, storage.ErrNotFound
	}
	return *user, nil
}

// GetByFirebaseUID get user by firebaseUID
func (s UserStorage) GetByFirebaseUID(firebaseUID string, context api2go.APIContexter) (model.User, error) {
	result := model.User{}
	for _, user := range s.users {
		if user.FirebaseUID == firebaseUID {
			return *user, nil
		}
	}
	return result, storage.ErrNotFound
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
		return storage.ErrNotFound
	}
	delete(s.users, id)

	return nil
}

// Update a user
func (s *UserStorage) Update(u model.User, context api2go.APIContexter) error {
	_, exists := s.users[u.GetID()]
	if !exists {
		return storage.ErrNotFound
	}
	s.users[u.GetID()] = &u
	return nil
}
