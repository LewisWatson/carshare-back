package resource

import (
	"errors"
	"fmt"
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

	carShares, err := cs.CarShareStorage.GetAll(r.Context)
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
	for _, carShare := range carShares {
		err = cs.populate(&carShare, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
			return &Response{Res: result}, api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
		result = append(result, carShare)
	}

	return &Response{Res: result}, nil
}

// FindOne carShare
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

// Create a new carShare
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

// Delete a carShare :(
func (cs CarShareResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {

	err := cs.CarShareStorage.Delete(id, r.Context)

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

	return &Response{Code: http.StatusOK}, nil
}

// Update a carShare
func (cs CarShareResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	carShare, ok := obj.(model.CarShare)

	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to car share update: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
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

// Populate the relationships for a car share
func (cs CarShareResource) populate(carShare *model.CarShare, context api2go.APIContexter) error {

	carShare.Trips = []model.Trip{}
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
