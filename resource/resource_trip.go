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
	TripStorage     *storage.TripStorage
	UserStorage     *storage.UserStorage
	CarShareStorage *storage.CarShareStorage
	ScoreStorage    *storage.ScoreStorage
}

// FindAll trips
func (t TripResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	var result []model.Trip
	for _, trip := range t.TripStorage.GetAll() {

		if trip.CarShare != nil {
			carShare, err := t.CarShareStorage.GetOne(trip.CarShare.GetID())
			if err != nil {
				return &Response{}, err
			}
			trip.CarShare = &carShare
		}

		if trip.Driver != nil {
			driver, err := t.UserStorage.GetOne(trip.Driver.GetID())
			if err != nil {
				return &Response{}, err
			}
			trip.Driver = &driver
		}

		for _, passengerID := range trip.PassengerIDs {
			passenger, err := t.UserStorage.GetOne(passengerID)
			if err != nil {
				return &Response{}, err
			}
			trip.Passengers = append(trip.Passengers, &passenger)
		}

		for _, scoreID := range trip.ScoreIDs {
			score, err := t.ScoreStorage.GetOne(scoreID)
			if err != nil {
				return &Response{}, err
			}
			trip.Scores = append(trip.Scores, &score)
		}

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

	if trip.CarShare != nil {
		carShare, err2 := t.CarShareStorage.GetOne(trip.CarShare.GetID())
		if err2 != nil {
			return &Response{}, err
		}
		trip.CarShare = &carShare
	}

	if trip.Driver != nil {
		driver, err3 := t.UserStorage.GetOne(trip.Driver.GetID())
		if err3 != nil {
			return &Response{}, err
		}
		trip.Driver = &driver
	}

	for _, passengerID := range trip.PassengerIDs {
		passenger, err4 := t.UserStorage.GetOne(passengerID)
		if err4 != nil {
			return &Response{}, err
		}
		trip.Passengers = append(trip.Passengers, &passenger)
	}

	for _, scoreID := range trip.ScoreIDs {
		score, err5 := t.ScoreStorage.GetOne(scoreID)
		if err5 != nil {
			return &Response{}, err
		}
		trip.Scores = append(trip.Scores, &score)
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
