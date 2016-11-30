package resource

import (
	"errors"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/benbjohnson/clock"
	"github.com/manyminds/api2go"
)

// TripResource for api2go routes
type TripResource struct {
	TripStorage     storage.TripStorage
	UserStorage     storage.UserStorage
	CarShareStorage storage.CarShareStorage
	Clock           clock.Clock
}

// FindAll trips
func (t TripResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var result []model.Trip

	trips, err := t.TripStorage.GetAll(r.Context)
	if err != nil {
		return &Response{}, err
	}

	for _, trip := range trips {

		if trip.CarShareID != "" {
			var carShare model.CarShare
			carShare, err = t.CarShareStorage.GetOne(trip.CarShareID)
			if err != nil {
				return &Response{}, err
			}
			trip.CarShare = &carShare
		}

		if trip.DriverID != "" {
			var driver model.User
			driver, err = t.UserStorage.GetOne(trip.DriverID, r.Context)
			if err != nil {
				return &Response{}, err
			}
			trip.Driver = &driver
		}

		for _, passengerID := range trip.PassengerIDs {
			var passenger model.User
			passenger, err = t.UserStorage.GetOne(passengerID, r.Context)
			if err != nil {
				return &Response{}, err
			}
			trip.Passengers = append(trip.Passengers, &passenger)
		}

		result = append(result, trip)
	}

	return &Response{Res: result}, nil
}

// FindOne trip
func (t TripResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {

	trip, err := t.TripStorage.GetOne(ID, r.Context)
	if err != nil {
		return &Response{}, err
	}

	if trip.CarShareID != "" {
		var carShare model.CarShare
		carShare, err = t.CarShareStorage.GetOne(trip.CarShareID)
		if err != nil {
			return &Response{}, err
		}
		trip.CarShare = &carShare
	}

	if trip.DriverID != "" {
		var driver model.User
		driver, err = t.UserStorage.GetOne(trip.DriverID, r.Context)
		if err != nil {
			return &Response{}, err
		}
		trip.Driver = &driver
	}

	for _, passengerID := range trip.PassengerIDs {
		var passenger model.User
		passenger, err = t.UserStorage.GetOne(passengerID, r.Context)
		if err != nil {
			return &Response{}, err
		}
		trip.Passengers = append(trip.Passengers, &passenger)
	}

	return &Response{Res: trip}, err
}

// Create a new trip
func (t TripResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	trip, ok := obj.(model.Trip)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	if trip.CarShareID != "" {
		carShare, err := t.CarShareStorage.GetOne(trip.CarShareID)
		if err != nil {
			return &Response{}, err
		}
		trip.CarShare = &carShare
	}

	if trip.DriverID != "" {
		driver, err := t.UserStorage.GetOne(trip.DriverID, r.Context)
		if err != nil {
			return &Response{}, err
		}
		trip.Driver = &driver
	}

	for _, passengerID := range trip.PassengerIDs {
		passenger, err := t.UserStorage.GetOne(passengerID, r.Context)
		if err != nil {
			return &Response{}, err
		}
		trip.Passengers = append(trip.Passengers, &passenger)
	}

	trip.Scores = make(map[string]model.Score)
	if trip.CarShareID != "" {
		latestTrip, err := t.TripStorage.GetLatest(trip.CarShareID, r.Context)
		if err != nil {
			return &Response{}, err
		}
		trip.CalculateScores(latestTrip.Scores)
	}

	trip.TimeStamp = t.Clock.Now().UTC()

	id, err := t.TripStorage.Insert(trip, r.Context)
	if err != nil {
		return &Response{}, err
	}

	trip.SetID(id)

	return &Response{Res: trip, Code: http.StatusCreated}, nil
}

// Delete a trip :(
func (t TripResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {
	err := t.TripStorage.Delete(id, r.Context)
	return &Response{Code: http.StatusOK}, err
}

// Update a trip
func (t TripResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	trip, ok := obj.(model.Trip)
	if !ok {
		return &Response{}, api2go.NewHTTPError(errors.New("Invalid instance given"), "Invalid instance given", http.StatusBadRequest)
	}

	if trip.CarShareID != "" {
		latestTrip, err := t.TripStorage.GetLatest(trip.CarShareID, r.Context)
		if err != nil {
			return &Response{}, err
		}
		trip.CalculateScores(latestTrip.Scores)
	}

	err := t.TripStorage.Update(trip, r.Context)
	return &Response{Res: trip, Code: http.StatusNoContent}, err
}
