package resource

import (
	"fmt"
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

// FindAll to satisfy api2go.FindAll interface
func (t TripResource) FindAll(r api2go.Request) (api2go.Responder, error) {
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
			return &Response{Res: trips}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
	}
	return &Response{Res: trips}, nil
}

// FindOne to satisfy api2go.CRUD interface
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

// Create to satisfy api2go.CRUD interface
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

	trip.TimeStamp = t.Clock.Now().UTC()

	id, err := t.TripStorage.Insert(trip, r.Context)
	if err != nil {
		errMsg := "Error occurred while persisting trip"
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	err = trip.SetID(id)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			err,
			err.Error(),
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

	return &Response{Res: trip, Code: http.StatusCreated}, err
}

// Delete to satisfy the api2go.CRUD interface
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

// Update to satisfy api2go.CRUD interface
func (t TripResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	trip, ok := obj.(model.Trip)
	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip update: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}
	if trip.CarShareID != "" {
		/*
		 * Prevent trips from being changing car shares after they are initially assigned
		 */
		tripInDataStore, err := t.TripStorage.GetOne(trip.GetID(), r.Context)
		if err != nil {
			if err == storage.ErrNotFound {
				return &Response{}, api2go.NewHTTPError(
					fmt.Errorf("Unable to find trip %s to update", trip.GetID()),
					http.StatusText(http.StatusNotFound),
					http.StatusNotFound,
				)
			}
			errMsg := fmt.Sprintf("Error retrieving trip %s", trip.GetID())
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
		if tripInDataStore.CarShareID != "" && tripInDataStore.CarShareID != trip.CarShareID {
			errMsg := fmt.Sprintf("trip %s already belongs to another car share", trip.GetID())
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s", errMsg),
				errMsg,
				http.StatusInternalServerError,
			)
		}

		// add trip to car shares trip list
		carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, r.Context)
		if err != nil {
			if err == storage.ErrNotFound {
				err = fmt.Errorf("Unable to find car share %s to in order to add trip relationship", trip.CarShareID)
				return &Response{}, api2go.NewHTTPError(
					err,
					err.Error(),
					http.StatusInternalServerError,
				)
			}
			errMsg := fmt.Sprintf("Error retrieving car share %s in order to add trip relationship", trip.CarShareID)
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
		found := false
		for _, tripID := range carShare.TripIDs {
			if tripID == trip.GetID() {
				found = true
				break
			}
		}
		if !found {
			carShare.TripIDs = append(carShare.TripIDs, trip.GetID())
			err = t.CarShareStorage.Update(carShare, r.Context)
			if err != nil {
				errMsg := fmt.Sprintf("Error updating car share %s", carShare.GetID())
				return &Response{}, api2go.NewHTTPError(
					fmt.Errorf("%s, %s", errMsg, err),
					errMsg,
					http.StatusInternalServerError,
				)
			}
		}
	}

	// verify driver
	if trip.DriverID != "" {
		_, err := t.UserStorage.GetOne(trip.DriverID, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error verifying driver %s", trip.DriverID)
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
	}

	// verify passengers
	for _, passengerID := range trip.PassengerIDs {
		if passengerID == trip.DriverID {
			err := fmt.Errorf("Error passenger %s is set as driver", passengerID)
			return &Response{}, api2go.NewHTTPError(
				err,
				err.Error(),
				http.StatusBadRequest,
			)
		}
		_, err := t.UserStorage.GetOne(passengerID, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error verifying passenger %s", passengerID)
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
	}

	// TODO recalculate scores for trips that occur after this one as well
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

	err = t.TripStorage.Update(trip, r.Context)

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

// populate the relationships for a trip
func (t TripResource) populate(trip *model.Trip, context api2go.APIContexter) error {

	trip.Driver = nil
	if trip.DriverID != "" {
		driver, err := t.UserStorage.GetOne(trip.DriverID, context)
		if err != nil {
			return err
		}
		trip.Driver = &driver
	}

	trip.Passengers = nil
	for _, passengerID := range trip.PassengerIDs {
		passenger, err := t.UserStorage.GetOne(passengerID, context)
		if err != nil {
			return err
		}
		trip.Passengers = append(trip.Passengers, &passenger)
	}

	return nil
}
