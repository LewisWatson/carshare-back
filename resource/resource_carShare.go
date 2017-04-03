package resource

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/LewisWatson/firebase-jwt-auth.v1"
)

// CarShareResource for api2go routes
type CarShareResource struct {
	CarShareStorage storage.CarShareStorage
	TripStorage     storage.TripStorage
	UserStorage     storage.UserStorage
	TokenVerifier   fireauth.TokenVerifier
}

var (

	/*
	 * Metrics we shall be gathering
	 */
	carShareFindAllDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "carshare_find_all_duration_seconds",
		Help: "Time taken to find all users",
	}, []string{"code"})
	carShareFindOneDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "carshare_find_one_duration_seconds",
		Help: "Time taken to find one users",
	}, []string{"code"})
	carShareCreateDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "carshare_create_duration_seconds",
		Help: "Time taken to create users",
	}, []string{"code"})
	carShareDeleteDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "carshare_delete_duration_seconds",
		Help: "Time taken to delete users",
	}, []string{"code"})
	carShareUpdateDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "carshare_update_duration_seconds",
		Help: "Time taken to update users",
	}, []string{"code"})
)

func init() {

	/*
	 * Register metric counters with prometheus
	 */
	prometheus.MustRegister(carShareFindAllDurationSeconds)
	prometheus.MustRegister(carShareFindOneDurationSeconds)
	prometheus.MustRegister(carShareCreateDurationSeconds)
	prometheus.MustRegister(carShareDeleteDurationSeconds)
	prometheus.MustRegister(carShareUpdateDurationSeconds)

}

// FindAll to satisfy api2go.FindAll interface
func (cs CarShareResource) FindAll(r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer tripFindAllDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, cs.TokenVerifier, cs.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(err, http.StatusText(code), code)
	}

	result, err := cs.CarShareStorage.GetAll(requestingUser.GetID(), r.Context)
	if err != nil {
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Error retrieving all car shares, %s", err),
			http.StatusText(code),
			code,
		)
	}

	// if an error occurs while populating, still attempt to send the remainder of the response. Don't store code for metrics
	for _, carShare := range result {
		err = cs.populate(&carShare, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
			return &Response{Res: result}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, http.StatusInternalServerError)
		}
	}

	code = http.StatusOK
	return &Response{Res: result, Code: code}, nil
}

// FindOne to satisfy api2go.CRUD interface
func (cs CarShareResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer tripFindOneDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, cs.TokenVerifier, cs.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Error retrieving car share, %s", err), http.StatusText(code), code)
	}

	carShare, err := cs.CarShareStorage.GetOne(ID, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("unable to find car share %s", ID), http.StatusText(code), code)
	default:
		errMsg := fmt.Sprintf("Error occurred while retrieving car share %s", ID)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	if !carShare.IsMember(requestingUser.GetID()) {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("User %v not member of car share %v", requestingUser.GetID(), carShare.GetID()), http.StatusText(code), code)
	}

	// if an error occurs while populating, still attempt to send the remainder of the response. Don't store code for metrics
	popErr := cs.populate(&carShare, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
		err = api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, http.StatusInternalServerError)
	}

	code = http.StatusOK
	return &Response{Res: carShare, Code: code}, err
}

// Create to satisfy api2go.CRUD interface
func (cs CarShareResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer tripCreateDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, cs.TokenVerifier, cs.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Error creating car share, %s", err), http.StatusText(code), code)
	}

	carShare, ok := obj.(model.CarShare)
	if !ok {
		code = http.StatusBadRequest
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to car share create: %v", obj), http.StatusText(code), code)
	}

	if !carShare.IsMember(requestingUser.GetID()) {
		carShare.MemberIDs = append(carShare.MemberIDs, requestingUser.GetID())
	}

	if !carShare.IsAdmin(requestingUser.GetID()) {
		carShare.AdminIDs = append(carShare.AdminIDs, requestingUser.GetID())
	}

	id, err := cs.CarShareStorage.Insert(carShare, r.Context)
	if err == nil && id == "" {
		err = errors.New("null id returned")
	}
	if err != nil {
		errMsg := "Error occurred while persisting car share"
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	carShare.SetID(id)

	// if an error occurs while populating, still attempt to send the remainder of the response. Don't store code for metrics
	popErr := cs.populate(&carShare, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
		err = api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, http.StatusInternalServerError)
	}

	code = http.StatusCreated
	return &Response{Res: carShare, Code: code}, err
}

// Delete to satisfy api2go.CRUD interface
func (cs CarShareResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer tripDeleteDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, cs.TokenVerifier, cs.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Error deleting car share, %s", err), http.StatusText(http.StatusForbidden), code)
	}

	carShare, err := cs.CarShareStorage.GetOne(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("unable to find car share %s", id), http.StatusText(code), code)
	default:
		errMsg := fmt.Sprintf("Error occurred while retrieving car share %s", id)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	if !carShare.IsAdmin(requestingUser.GetID()) {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Non admin user %v attempting to delete car share %v", requestingUser.GetID(), carShare.GetID()),
			http.StatusText(code),
			code,
		)
	}

	err = cs.CarShareStorage.Delete(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("unable to find car share %s to delete", id), http.StatusText(code), code)
	default:
		errMsg := fmt.Sprintf("Error occurred while deleting car share %s", id)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	ok := cs.deleteAssocTrips(carShare, r.Context)
	if !ok {
		errMsg := fmt.Sprintf("Car share deleted, but error occurred while deleting associated trips")
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s", errMsg), errMsg, code)
	}

	code = http.StatusOK
	return &Response{Code: code}, nil
}

func (cs CarShareResource) deleteAssocTrips(carShare model.CarShare, ctx api2go.APIContexter) bool {
	ok := true
	for _, tripID := range carShare.TripIDs {
		err := cs.TripStorage.Delete(tripID, ctx)
		if err != nil && err != storage.ErrNotFound {
			ok = false
			log.Printf("Error deleting associated trip %s, %v", tripID, err)
		}
	}
	return ok
}

// Update to satisfy api2go.CRUD interface
func (cs CarShareResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer tripUpdateDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, cs.TokenVerifier, cs.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Error updating car share, %s", err), http.StatusText(code), code)
	}

	carShare, ok := obj.(model.CarShare)
	if !ok {
		code = http.StatusBadRequest
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Invalid instance given to car share update: %v", obj), http.StatusText(code), code)
	}

	existingCarShare, err := cs.CarShareStorage.GetOne(carShare.GetID(), r.Context)
	if err != nil {
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Unable to find carShare %v", carShare.GetID()), http.StatusText(code), code)
	}

	if !existingCarShare.IsAdmin(requestingUser.GetID()) {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Non admin user %v attempting to update carShare %v", requestingUser.GetID(), carShare.GetID()),
			http.StatusText(code),
			code,
		)
	}

	// verify that tripID's link to real trips, and if required set those trips as belonging to this car share
	for _, tripID := range carShare.TripIDs {

		trip, err := cs.TripStorage.GetOne(tripID, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error verifying trip %s", tripID)
			code = http.StatusInternalServerError
			return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
		}

		if trip.CarShareID == "" {
			trip.CarShareID = carShare.GetID()
			err = cs.TripStorage.Update(trip, r.Context)
			if err != nil {
				errMsg := fmt.Sprintf("Error occured while assigning trip %s to car share %s", tripID, carShare.GetID())
				code = http.StatusInternalServerError
				return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
			}
			log.Printf("trip %s updated to belong to car share %s", trip.GetID(), carShare.GetID())
		}

		// do not allow trips to be transferred between car shares as that doesn't make sense
		if trip.CarShareID != carShare.GetID() {
			errMsg := fmt.Sprintf("trip %s already belongs to another car share", tripID)
			code = http.StatusInternalServerError
			return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s", errMsg), errMsg, code)
		}

	}

	// verify that admins link to actual users
	for _, adminID := range carShare.AdminIDs {
		_, err := cs.UserStorage.GetOne(adminID, r.Context)
		if err != nil {
			errMsg := fmt.Sprintf("Error verifying user %s", adminID)
			code = http.StatusInternalServerError
			return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
		}
	}

	err = cs.CarShareStorage.Update(carShare, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Unable to find car share %s to update", carShare.GetID()), http.StatusText(code), code)
	default:
		errMsg := fmt.Sprintf("Error occurred while updating car share %s", carShare.GetID())
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	// if an error occurs while populating, still attempt to send the remainder of the response. Don't store code for metrics
	popErr := cs.populate(&carShare, r.Context)
	if popErr != nil {
		errMsg := fmt.Sprintf("Error when populating car share %s", carShare.GetID())
		err = api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, http.StatusInternalServerError)
	}

	code = http.StatusNoContent
	return &Response{Res: carShare, Code: code}, err
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

	carShare.Members = nil
	for _, memberID := range carShare.MemberIDs {
		member, err := cs.UserStorage.GetOne(memberID, context)
		if err != nil {
			return err
		}
		carShare.Members = append(carShare.Members, &member)
	}

	return nil
}
