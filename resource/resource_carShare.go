package resource

import (
	"errors"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// CarShareResource for api2go routes
type CarShareResource struct {
	CarShareStorage storage.CarShareStorage
	TripStorage     storage.TripStorage
	UserStorage     storage.UserStorage
}

// FindAll carShares
func (cs CarShareResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var result []model.CarShare

	carShares, err := cs.CarShareStorage.GetAll()

	if err != nil {
		return &Response{}, err
	}

	for _, carShare := range carShares {
		// get all trips for the carShare
		carShare.Trips = []*model.Trip{}

		for _, tripID := range carShare.TripIDs {
			trip, err := cs.TripStorage.GetOne(tripID)
			if err != nil {
				return &Response{}, err
			}
			carShare.Trips = append(carShare.Trips, &trip)
		}

		for _, adminID := range carShare.AdminIDs {
			admin, err := cs.UserStorage.GetOne(adminID)
			if err != nil {
				return &Response{}, err
			}
			carShare.Admins = append(carShare.Admins, &admin)
		}

		result = append(result, carShare)
	}

	return &Response{Res: result}, nil
}

// FindOne carShare
func (cs CarShareResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {

	var err error

	carShare, err := cs.CarShareStorage.GetOne(ID)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(err, err.Error(), http.StatusNotFound)
	}
	// get all trips for the carShare
	carShare.Trips = []*model.Trip{}
	for _, tripID := range carShare.TripIDs {
		var trip model.Trip
		trip, err = cs.TripStorage.GetOne(tripID)
		if err != nil {
			return &Response{}, err
		}
		carShare.Trips = append(carShare.Trips, &trip)
	}

	for _, adminID := range carShare.AdminIDs {
		var admin model.User
		admin, err = cs.UserStorage.GetOne(adminID)
		if err != nil {
			return &Response{}, err
		}
		carShare.Admins = append(carShare.Admins, &admin)
	}
	return &Response{Res: carShare}, err
}

// Create a new carShare
func (cs CarShareResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	carShare, ok := obj.(model.CarShare)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	id, err := cs.CarShareStorage.Insert(carShare)
	if err != nil {
		return &Response{}, err
	}

	carShare.SetID(id)
	return &Response{Res: carShare, Code: http.StatusCreated}, nil
}

// Delete a carShare :(
func (cs CarShareResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {
	err := cs.CarShareStorage.Delete(id)
	return &Response{Code: http.StatusOK}, err
}

// Update a carShare
func (cs CarShareResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	carShare, ok := obj.(model.CarShare)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	err := cs.CarShareStorage.Update(carShare)
	return &Response{Res: carShare, Code: http.StatusNoContent}, err
}
