package resource

import (
	"fmt"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/benbjohnson/clock"
	"github.com/manyminds/api2go"
	"gopkg.in/LewisWatson/firebase-jwt-auth.v1"
)

// TripResource for api2go routes
type TripResource struct {
	TripStorage     storage.TripStorage
	UserStorage     storage.UserStorage
	CarShareStorage storage.CarShareStorage
	TokenVerifier   fireauth.TokenVerifier
	Clock           clock.Clock
}

// FindAll to satisfy api2go.FindAll interface
func (t TripResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	return &Response{}, api2go.NewHTTPError(
		fmt.Errorf("Find all trips not supported"),
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
}

// FindOne to satisfy api2go.CRUD interface
func (t TripResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {

	requestingUser, err := getRequestUser(r, t.TokenVerifier, t.UserStorage)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			err,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden,
		)
	}

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

	httpErr := t.checkUserIsMemberOfCarShare(requestingUser, trip, r.Context)
	if httpErr.Errors == nil {
		return &Response{}, httpErr
	}

	carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, r.Context)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find car share %s linked to trip %s: %s", trip.CarShareID, trip.GetID(), err),
			"Error finding associated car share",
			http.StatusInternalServerError,
		)
	}

	if !carShare.IsMember(requestingUser.GetID()) && !carShare.IsAdmin(requestingUser.GetID()) {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("user %s attempting to access trip %s for which they are not a member or admin for", requestingUser.GetID(), trip.GetID()),
			"must be a member or admin for associated carshare to access",
			http.StatusForbidden,
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

	requestingUser, err := getRequestUser(r, t.TokenVerifier, t.UserStorage)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			err,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden,
		)
	}

	trip, ok := obj.(model.Trip)
	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip create: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}

	httpErr := t.checkUserIsMemberOfCarShare(requestingUser, trip, r.Context)
	if httpErr.Errors == nil {
		return &Response{}, httpErr
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

	requestingUser, err := getRequestUser(r, t.TokenVerifier, t.UserStorage)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			err,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden,
		)
	}

	trip, err := t.TripStorage.GetOne(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find trip %s", id),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while retrieving trip %s", id)
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	httpErr := t.checkUserIsMemberOfCarShare(requestingUser, trip, r.Context)
	if httpErr.Errors == nil {
		return &Response{}, httpErr
	}

	err = t.TripStorage.Delete(id, r.Context)
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

	carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, r.Context)
	switch err {
	case nil:
	case storage.ErrNotFound:
		break
	default:
		errMsg := fmt.Sprintf("Trip deleted but error occurred while retrieving car share %s associated with trip %s", trip.CarShareID, id)
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}
	for index := range carShare.TripIDs {
		if carShare.TripIDs[index] == id {
			carShare.TripIDs = append(carShare.TripIDs[:index], carShare.TripIDs[index+1:]...)
			break
		}
	}
	t.CarShareStorage.Update(carShare, r.Context)

	return &Response{Code: http.StatusOK}, nil
}

// Update to satisfy api2go.CRUD interface
func (t TripResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	requestingUser, err := getRequestUser(r, t.TokenVerifier, t.UserStorage)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			err,
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden,
		)
	}

	trip, ok := obj.(model.Trip)
	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip update: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}

	if trip.CarShareID == "" {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip update (missing carShareID): %v", obj),
			"must provide a carShareID",
			http.StatusBadRequest,
		)
	}

	tripInDataStore, err := t.TripStorage.GetOne(trip.GetID(), r.Context)
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
		errMsg := fmt.Sprintf("Error retrieving trip %s", trip.GetID())
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	// important to check against the trip in the data store
	httpErr := t.checkUserIsMemberOfCarShare(requestingUser, tripInDataStore, r.Context)
	if httpErr.Errors == nil {
		return &Response{}, httpErr
	}

	// Prevent trips from being re-assigned car shares
	if tripInDataStore.CarShareID != "" && tripInDataStore.CarShareID != trip.CarShareID {
		errMsg := fmt.Sprintf("trip %s already belongs to another car share", trip.GetID())
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s", errMsg),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	httpErr = t.addToCarShareTripList(trip, r.Context)
	if httpErr.Errors != nil {
		return &Response{}, httpErr
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

// addToCarShareTripList updates the associated carshare and ensures that it has the trip in its list of trips
func (t TripResource) addToCarShareTripList(trip model.Trip, ctx api2go.APIContexter) api2go.HTTPError {

	carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, ctx)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		err = fmt.Errorf("Unable to find car share %s to in order to add trip relationship", trip.CarShareID)
		return api2go.NewHTTPError(
			err,
			err.Error(),
			http.StatusInternalServerError,
		)
	default:
		errMsg := fmt.Sprintf("Error retrieving car share %s in order to add trip relationship", trip.CarShareID)
		return api2go.NewHTTPError(
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
		err = t.CarShareStorage.Update(carShare, ctx)
		if err != nil {
			errMsg := fmt.Sprintf("Error updating car share %s", carShare.GetID())
			return api2go.NewHTTPError(
				fmt.Errorf("%s, %s", errMsg, err),
				errMsg,
				http.StatusInternalServerError,
			)
		}
	}

	return api2go.HTTPError{}
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

func (t TripResource) checkUserIsMemberOfCarShare(user model.User, trip model.Trip, ctx api2go.APIContexter) api2go.HTTPError {

	carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, ctx)
	if err != nil {
		return api2go.NewHTTPError(
			fmt.Errorf("unable to find car share %s linked to trip %s: %s", trip.CarShareID, trip.GetID(), err),
			"Error finding associated car share",
			http.StatusInternalServerError,
		)
	}

	if !carShare.IsMember(user.GetID()) {
		return api2go.NewHTTPError(
			fmt.Errorf("user %s attempting to access trip %s for car share %s they are not a member of", user.GetID(), trip.GetID(), trip.CarShareID),
			"must be a member or admin for associated carshare to access",
			http.StatusForbidden,
		)
	}

	return api2go.HTTPError{}
}
