package resource

import (
	"errors"
	"fmt"
	"log"
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

// FindAll to satisfy api2go.FindAll interface
func (cs CarShareResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	result, err := cs.CarShareStorage.GetAll(r.Context)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Error retrieveing all car shares, %s", err),
			http.StatusText(http.StatusInternalServerError),
			http.StatusInternalServerError,
		)
	}
	/*
	 * Populate the car share relationships. If an error occurs while we do this
	 * then return the error along with what has been retrieved up to that point
	 */
	for _, carShare := range result {
		err = cs.populate(&carShare, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
			return &Response{Res: result}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
	}

	return &Response{Res: result}, nil
}

// FindOne to satisfy api2go.CRUD interface
func (cs CarShareResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	carShare, err := cs.CarShareStorage.GetOne(ID, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find car share %s", ID),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while retrieving car share %s", ID)
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}
	// if an error occurs while populating, still attempt to send the remainder
	// of the response
	popErr := cs.populate(&carShare, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
		err = api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}
	return &Response{Res: carShare}, err
}

// Create to satisfy api2go.CRUD interface
func (cs CarShareResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {
	carShare, ok := obj.(model.CarShare)
	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to car share create: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}
	id, err := cs.CarShareStorage.Insert(carShare, r.Context)
	if err == nil && id == "" {
		err = errors.New("null id returned")
	}
	if err != nil {
		errMsg := "Error occurred while persisting car share"
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}
	carShare.SetID(id)
	// if an error occurs while populating, still attempt to send the remainder
	// of the response
	popErr := cs.populate(&carShare, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
		err = api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}
	return &Response{Res: carShare, Code: http.StatusCreated}, err
}

// Delete to satisfy api2go.CRUD interface
func (cs CarShareResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {

	carShare, err := cs.CarShareStorage.GetOne(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find car share %s", id),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while retrieving car share %s", id)
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	err = cs.CarShareStorage.Delete(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find car share %s to delete", id),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while deleting car share %s", id)
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	var tripsDeletedOK = true
	for _, tripID := range carShare.TripIDs {
		err = cs.TripStorage.Delete(tripID, r.Context)
		if err != nil && err != storage.ErrNotFound {
			tripsDeletedOK = false
			log.Printf("Error deleting associated trip %s, %v", tripID, err)
		}
	}

	if !tripsDeletedOK {
		errMsg := fmt.Sprintf("Car share deleted, but error occurred while deleting associated trips")
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s", errMsg),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	return &Response{Code: http.StatusOK}, nil
}

// Update to satisfy api2go.CRUD interface
func (cs CarShareResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	carShare, ok := obj.(model.CarShare)
	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to car share update: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}

	// verify that tripID's link to real trips, and if required set those trips as belonging to this car share
	for _, tripID := range carShare.TripIDs {

		trip, err := cs.TripStorage.GetOne(tripID, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error verifying trip %s", tripID)
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}

		if trip.CarShareID == "" {
			trip.CarShareID = carShare.GetID()
			err = cs.TripStorage.Update(trip, r.Context)
			if err != nil {
				errMsg := fmt.Sprintf("Error occured while assigning trip %s to car share %s", tripID, carShare.GetID())
				return &Response{}, api2go.NewHTTPError(
					fmt.Errorf("%s, %s", errMsg, err),
					errMsg,
					http.StatusInternalServerError,
				)
			}
			log.Printf("trip %s updated to belong to car share %s", trip.GetID(), carShare.GetID())
		}

		// do not allow trips to be transferred between car shares as that doesn't make sense
		if trip.CarShareID != carShare.GetID() {
			errMsg := fmt.Sprintf("trip %s already belongs to another car share", tripID)
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s", errMsg),
				errMsg,
				http.StatusInternalServerError,
			)
		}

	}

	// verify that admins link to actual users
	for _, adminID := range carShare.AdminIDs {
		_, err := cs.UserStorage.GetOne(adminID, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error verifying user %s", adminID)
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
	}

	err := cs.CarShareStorage.Update(carShare, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Unable to find car share %s to update", carShare.GetID()),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while updating car share %s", carShare.GetID())
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	popErr := cs.populate(&carShare, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
		err = api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	return &Response{Res: carShare, Code: http.StatusNoContent}, err
}

// populate the relationships for a car share
func (cs CarShareResource) populate(carShare *model.CarShare, context api2go.APIContexter) error {

	carShare.Trips = nil
	for _, tripID := range carShare.TripIDs {
		trip, err := cs.TripStorage.GetOne(tripID, context)
		if err != nil {
			return err
		}
		carShare.Trips = append(carShare.Trips, trip)
	}

	carShare.Admins = nil
	for _, adminID := range carShare.AdminIDs {
		admin, err := cs.UserStorage.GetOne(adminID, context)
		if err != nil {
			return err
		}
		carShare.Admins = append(carShare.Admins, &admin)
	}

	return nil
}
