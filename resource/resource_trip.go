package resource

import (
	"fmt"
	"net/http"
	"time"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/LewisWatson/firebase-jwt-auth"
	"github.com/benbjohnson/clock"
	"github.com/manyminds/api2go"
	"github.com/prometheus/client_golang/prometheus"
)

// TripResource for api2go routes
type TripResource struct {
	TripStorage     storage.TripStorage
	UserStorage     storage.UserStorage
	CarShareStorage storage.CarShareStorage
	TokenVerifier   fireauth.TokenVerifier
	Clock           clock.Clock
}

var (

	/*
	 * Metrics we shall be gathering
	 */
	tripFindAllDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trip_find_all_duration_seconds",
		Help: "Time taken to find all trips",
	}, []string{"code"})
	tripFindOneDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trip_find_one_duration_seconds",
		Help: "Time taken to find one trips",
	}, []string{"code"})
	tripCreateDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trip_create_duration_seconds",
		Help: "Time taken to create trips",
	}, []string{"code"})
	tripDeleteDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trip_delete_duration_seconds",
		Help: "Time taken to delete trips",
	}, []string{"code"})
	tripUpdateDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "trip_update_duration_seconds",
		Help: "Time taken to update trips",
	}, []string{"code"})
)

func init() {

	/*
	 * Register metric counters with prometheus
	 */
	prometheus.MustRegister(tripFindAllDurationSeconds)
	prometheus.MustRegister(tripFindOneDurationSeconds)
	prometheus.MustRegister(tripCreateDurationSeconds)
	prometheus.MustRegister(tripDeleteDurationSeconds)
	prometheus.MustRegister(tripUpdateDurationSeconds)

}

// FindAll to satisfy api2go.FindAll interface
func (t TripResource) FindAll(r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusMethodNotAllowed
	defer tripFindAllDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	return &Response{}, api2go.NewHTTPError(
		fmt.Errorf("Find all trips not supported"),
		http.StatusText(code),
		code,
	)
}

// FindOne to satisfy api2go.CRUD interface
func (t TripResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer tripFindOneDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, t.TokenVerifier, t.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(err, http.StatusText(code), code)
	}

	trip, err := t.TripStorage.GetOne(ID, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("unable to find trip %s", ID), http.StatusText(code), code)
	default:
		errMsg := fmt.Sprintf("Error occurred while retrieving trip %s", ID)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	httpErr, code := t.verifyCarShareMember(requestingUser, trip, r.Context)
	if httpErr.Errors != nil {
		return &Response{}, httpErr
	}

	carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, r.Context)
	if err != nil {
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find car share %s linked to trip %s: %s", trip.CarShareID, trip.GetID(), err),
			"Error finding associated car share",
			code,
		)
	}

	if !carShare.IsMember(requestingUser.GetID()) && !carShare.IsAdmin(requestingUser.GetID()) {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("user %s attempting to access trip %s for which they are not a member or admin for", requestingUser.GetID(), trip.GetID()),
			"must be a member or admin for associated carshare to access",
			code,
		)
	}

	// if an error occurs while populating, still attempt to send the remainder of the response. Don't store code for metrics
	popErr := t.populate(&trip, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating trip %s", trip.GetID())
		err = api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, http.StatusInternalServerError)
	}

	code = http.StatusOK
	return &Response{Res: trip}, err
}

// Create to satisfy api2go.CRUD interface
func (t TripResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer tripCreateDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, t.TokenVerifier, t.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(err, http.StatusText(code), code)
	}

	trip, ok := obj.(model.Trip)
	if !ok {
		code = http.StatusBadRequest
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip create: %v", obj),
			http.StatusText(code),
			code,
		)
	}

	if trip.CarShareID == "" {
		code = http.StatusBadRequest
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip create (missing carShareID): %v", obj),
			"must provide a carShareID",
			code,
		)
	}

	httpErr, code := t.verifyCarShareMember(requestingUser, trip, r.Context)
	if httpErr.Errors != nil {
		return &Response{}, httpErr
	}

	trip.Scores = make(map[string]model.Score)

	// TODO make custom store method to just return the scores
	latestTrip, err := t.TripStorage.GetLatest(trip.CarShareID, r.Context)
	if err != nil && err != storage.ErrNotFound {
		errMsg := fmt.Sprintf("Error retrieving latest trip for car share %s", trip.CarShareID)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			code,
		)
	}
	trip.CalculateScores(latestTrip.Scores)

	trip.TimeStamp = t.Clock.Now().UTC()

	id, err := t.TripStorage.Insert(trip, r.Context)
	if err != nil {
		errMsg := "Error occurred while persisting trip"
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			code,
		)
	}

	err = trip.SetID(id)
	if err != nil {
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(err, err.Error(), code)
	}

	// if an error occurs while populating, still attempt to send the remainder of the response. Don't store code for metrics
	popErr := t.populate(&trip, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating trip %s", trip.GetID())
		err = api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, http.StatusInternalServerError)
	}

	code = http.StatusCreated
	return &Response{Res: trip, Code: code}, err
}

// Delete to satisfy the api2go.CRUD interface
func (t TripResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer tripDeleteDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, t.TokenVerifier, t.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(err, http.StatusText(code), code)
	}

	trip, err := t.TripStorage.GetOne(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find trip %s", id),
			http.StatusText(code),
			code,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while retrieving trip %s", id)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			code,
		)
	}

	httpErr, code := t.verifyCarShareMember(requestingUser, trip, r.Context)
	if httpErr.Errors != nil {
		return &Response{}, httpErr
	}

	err = t.TripStorage.Delete(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find trip %s to delete", id),
			http.StatusText(code),
			code,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while deleting trip %s", id)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			code,
		)
	}

	carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, r.Context)
	switch err {
	case nil:
	case storage.ErrNotFound:
		break
	default:
		errMsg := fmt.Sprintf("Trip deleted but error occurred while retrieving car share %s associated with trip %s", trip.CarShareID, id)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}
	for index := range carShare.TripIDs {
		if carShare.TripIDs[index] == id {
			carShare.TripIDs = append(carShare.TripIDs[:index], carShare.TripIDs[index+1:]...)
			break
		}
	}
	t.CarShareStorage.Update(carShare, r.Context)

	code = http.StatusOK
	return &Response{Code: code}, nil
}

// Update to satisfy api2go.CRUD interface
func (t TripResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusMethodNotAllowed
	defer tripUpdateDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, t.TokenVerifier, t.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(err, http.StatusText(code), code)
	}

	trip, ok := obj.(model.Trip)
	if !ok {
		code = http.StatusBadRequest
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip update: %v", obj),
			http.StatusText(code),
			code,
		)
	}

	tripInDataStore, err := t.TripStorage.GetOne(trip.GetID(), r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Unable to find trip %s to update", trip.GetID()),
			http.StatusText(code),
			code,
		)
	default:
		errMsg := fmt.Sprintf("Error retrieving trip %s", trip.GetID())
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	// Prevent trips from being re-assigned car shares
	if tripInDataStore.CarShareID != trip.CarShareID {
		errMsg := fmt.Sprintf("trip %s already belongs to another car share", trip.GetID())
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s", errMsg),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	// important to check against the trip in the data store
	httpErr, code := t.verifyCarShareMember(requestingUser, tripInDataStore, r.Context)
	if httpErr.Errors != nil {
		return &Response{}, httpErr
	}

	httpErr, code = t.addToCarShareTripList(trip, r.Context)
	if httpErr.Errors != nil {
		return &Response{}, httpErr
	}

	// verify driver
	if trip.DriverID != "" {
		_, err := t.UserStorage.GetOne(trip.DriverID, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error verifying driver %s", trip.DriverID)
			code = http.StatusInternalServerError
			return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
		}
	}

	// verify passengers
	for _, passengerID := range trip.PassengerIDs {
		if passengerID == trip.DriverID {
			err := fmt.Errorf("Error passenger %s is set as driver", passengerID)
			code = http.StatusBadRequest
			return &Response{}, api2go.NewHTTPError(err, err.Error(), code)
		}
		_, err := t.UserStorage.GetOne(passengerID, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error verifying passenger %s", passengerID)
			code = http.StatusInternalServerError
			return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
		}
	}

	// TODO recalculate scores for trips that occur after this one as well
	latestTrip, err := t.TripStorage.GetLatest(trip.CarShareID, r.Context)
	if err != nil && err != storage.ErrNotFound {
		errMsg := fmt.Sprintf("Error retrieving latest trip for car share %s", trip.CarShareID)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}
	trip.CalculateScores(latestTrip.Scores)

	err = t.TripStorage.Update(trip, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Unable to find trip %s to update", trip.GetID()),
			http.StatusText(code),
			code,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while updating trip %s", trip.GetID())
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			code,
		)
	}

	// if an error occurs while populating, still attempt to send the remainder of the response. Don't store code for metrics
	popErr := t.populate(&trip, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating trip %s", trip.GetID())
		err = api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, http.StatusInternalServerError)
	}

	code = http.StatusNoContent
	return &Response{Res: trip, Code: code}, err
}

// addToCarShareTripList updates the associated carshare and ensures that it has the trip in its list of trips
func (t TripResource) addToCarShareTripList(trip model.Trip, ctx api2go.APIContexter) (httpErr api2go.HTTPError, code int) {

	carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, ctx)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		err = fmt.Errorf("Unable to find car share %s to in order to add trip relationship", trip.CarShareID)
		code = http.StatusInternalServerError
		return api2go.NewHTTPError(err, err.Error(), code), code
	default:
		errMsg := fmt.Sprintf("Error retrieving car share %s in order to add trip relationship", trip.CarShareID)
		code = http.StatusInternalServerError
		return api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code), code
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
			code = http.StatusInternalServerError
			return api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code), code
		}
	}

	return api2go.HTTPError{}, code
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

// verifyCarShareMember will return an error if the supplied user is not a member of the car share associated with the provided trip.
func (t TripResource) verifyCarShareMember(user model.User, trip model.Trip, ctx api2go.APIContexter) (httpErr api2go.HTTPError, code int) {

	carShare, err := t.CarShareStorage.GetOne(trip.CarShareID, ctx)
	if err != nil {
		code = http.StatusInternalServerError
		return api2go.NewHTTPError(
			fmt.Errorf("unable to find car share %s linked to trip %s: %s", trip.CarShareID, trip.GetID(), err),
			"Error finding associated car share",
			code,
		), code
	}

	if !carShare.IsMember(user.GetID()) {
		code = http.StatusForbidden
		return api2go.NewHTTPError(
			fmt.Errorf("user %s attempting to access trip %s for car share %s they are not a member of", user.GetID(), trip.GetID(), trip.CarShareID),
			"must be a member or admin for associated carshare to access",
			code,
		), code
	}

	return api2go.HTTPError{}, code
}
