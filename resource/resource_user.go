package resource

import (
	"errors"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// UserResource for api2go routes
type UserResource struct {
	UserStorage storage.UserStorage
}

// FindAll users
func (u UserResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	users, err := u.UserStorage.GetAll()
	return &Response{Res: users}, err
}

// FindOne user
func (u UserResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	res, err := u.UserStorage.GetOne(ID)
	return &Response{Res: res}, err
}

// Create a new user
func (u UserResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	user, ok := obj.(model.User)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	id, err := u.UserStorage.Insert(user)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	user.SetID(id)
	return &Response{Res: user, Code: http.StatusCreated}, nil
}

// Delete a user :(
func (u UserResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {
	err := u.UserStorage.Delete(id)
	return &Response{Code: http.StatusOK}, err
}

// Update a user
func (u UserResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	user, ok := obj.(model.User)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	err := u.UserStorage.Update(user)
	return &Response{Res: user, Code: http.StatusNoContent}, err
}
