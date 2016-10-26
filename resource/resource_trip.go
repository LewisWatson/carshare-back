package resource

import (
	"errors"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// TripResource for api2go routes
type TripResource struct {
	TripStorage *storage.TripStorage
	UserStorage *storage.UserStorage
}

// FindAll trips
func (t TripResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var result []model.Trip
	for _, trip := range t.TripStorage.GetAll() {
		user, err := t.UserStorage.GetOne(trip.UserID)
		if err != nil {
			return &Response{}, err
		}
		trip.User = &user
		result = append(result, trip)
	}

	return &Response{Res: result}, nil
}

// FindOne trip
func (t TripResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	trip, err := t.TripStorage.GetOne(ID)
	if err != nil {
		return &Response{}, err
	}
	if trip.UserID != "" {
		user, err2 := t.UserStorage.GetOne(trip.UserID)
		if err2 != nil {
			return &Response{}, err2
		}
		trip.User = &user
	}
	return &Response{Res: trip}, err
}

// Create a new trip
func (t TripResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	trip, ok := obj.(model.Trip)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	// trip.TimeStamp = time.Now()
	id := t.TripStorage.Insert(trip)
	trip.ID = id
	return &Response{Res: trip, Code: http.StatusCreated}, nil
}

// Delete a trip :(
func (t TripResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {
	err := t.TripStorage.Delete(id)
	return &Response{Code: http.StatusOK}, err
}

// Update a trip
func (t TripResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	trip, ok := obj.(model.Trip)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	err := t.TripStorage.Update(trip)
	return &Response{Res: trip, Code: http.StatusNoContent}, err
}
