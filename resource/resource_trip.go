package resource

import (
	"errors"
	"fmt"
	"log"
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
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Error retrieveing all trips, %s", err),
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}

	/*
	 * Populate the trip relationships. If an error occurs then return the error
	 * along with what has been retrieved up to that point
	 */
	for _, trip := range trips {
		err = t.populate(&trip, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error when populating trip %s", trip.GetID())
			return &Response{Res: result}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
		result = append(result, trip)
	}

	return &Response{Res: result}, nil
}

// FindOne trip
func (t TripResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {

	trip, err := t.TripStorage.GetOne(ID, r.Context)

	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find trip %s", ID),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while retrieving trip %s", ID)
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	// if an error occurs while populating, still attempt to send the remainder
	// of the response
	popErr := t.populate(&trip, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating trip %s", trip.GetID())
		err = api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	return &Response{Res: trip}, err
}

// Create a new trip
func (t TripResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	trip, ok := obj.(model.Trip)
	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip create: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}

	trip.Scores = make(map[string]model.Score)
	if trip.CarShareID != "" {
		// TODO make custom store method to just return the scores
		latestTrip, err := t.TripStorage.GetLatest(trip.CarShareID, r.Context)
		if err != nil && err != storage.ErrNotFound {
			errMsg := fmt.Sprintf("Error retrieving latest trip for car share %s", trip.CarShareID)
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
		trip.CalculateScores(latestTrip.Scores)
	}

	trip.TimeStamp = t.Clock.Now().UTC()

	id, err := t.TripStorage.Insert(trip, r.Context)
	if err == nil && id == "" {
		err = errors.New("null id returned")
	}
	if err != nil {
		errMsg := "Error occurred while persisting trip"
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	trip.SetID(id)

	// if an error occurs while populating, still attempt to send the remainder
	// of the response
	popErr := t.populate(&trip, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating trip %s", trip.GetID())
		err = api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	return &Response{Res: trip, Code: http.StatusCreated}, err
}

// Delete a trip :(
func (t TripResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {

	err := t.TripStorage.Delete(id, r.Context)

	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find trip %s to delete", id),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while deleting trip %s", id)
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	return &Response{Code: http.StatusOK}, nil

}

// Update a trip
func (t TripResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	trip, ok := obj.(model.Trip)

	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip update: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}

	log.Printf("Update trip %v", obj)

	if trip.CarShareID != "" {
		latestTrip, err := t.TripStorage.GetLatest(trip.CarShareID, r.Context)
		if err != nil && err != storage.ErrNotFound {
			errMsg := fmt.Sprintf("Error retrieving latest trip for car share %s", trip.CarShareID)
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
		trip.CalculateScores(latestTrip.Scores)
	}

	err := t.TripStorage.Update(trip, r.Context)

	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Unable to find trip %s to update", trip.GetID()),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while updating trip %s", trip.GetID())
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	// if an error occurs while populating, still attempt to send the remainder
	// of the response
	popErr := t.populate(&trip, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating trip %s", trip.GetID())
		err = api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	return &Response{Res: trip, Code: http.StatusNoContent}, err
}

// Populate the relationships for a trip
func (t TripResource) populate(trip *model.Trip, context api2go.APIContexter) error {

	trip.CarShare = nil
	if trip.CarShareID != "" {
		var carShare model.CarShare
		carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, context)
		if err != nil {
			return err
		}
		trip.CarShare = &carShare
	}

	trip.Driver = nil
	if trip.DriverID != "" {
		var driver model.User
		driver, err := t.UserStorage.GetOne(trip.DriverID, context)
		if err != nil {
			return err
		}
		trip.Driver = &driver
	}

	trip.Passengers = nil
	for _, passengerID := range trip.PassengerIDs {
		var passenger model.User
		passenger, err := t.UserStorage.GetOne(passengerID, context)
		if err != nil {
			return err
		}
		trip.Passengers = append(trip.Passengers, &passenger)
	}

	return nil
}
