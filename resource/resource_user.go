package resource

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
	"github.com/prometheus/client_golang/prometheus"
	"gopkg.in/LewisWatson/firebase-jwt-auth.v1"
)

// UserResource for api2go routes
type UserResource struct {
	UserStorage     storage.UserStorage
	CarShareStorage storage.CarShareStorage
	TokenVerifier   fireauth.TokenVerifier
}

var (

	/*
	 * Metrics we shall be gathering
	 */
	userFindAllDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "user_find_all_duration_seconds",
		Help: "Time taken to find all users",
	}, []string{"code"})
	userFindOneDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "user_find_one_duration_seconds",
		Help: "Time taken to find one users",
	}, []string{"code"})
	userCreateDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "user_create_duration_seconds",
		Help: "Time taken to create users",
	}, []string{"code"})
	userDeleteDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "user_delete_duration_seconds",
		Help: "Time taken to delete users",
	}, []string{"code"})
	userUpdateDurationSeconds = prometheus.NewHistogramVec(prometheus.HistogramOpts{
		Name: "user_update_duration_seconds",
		Help: "Time taken to update users",
	}, []string{"code"})
)

func init() {

	/*
	 * Register metric counters with prometheus
	 */
	prometheus.MustRegister(userFindAllDurationSeconds)
	prometheus.MustRegister(userFindOneDurationSeconds)
	prometheus.MustRegister(userCreateDurationSeconds)
	prometheus.MustRegister(userDeleteDurationSeconds)
	prometheus.MustRegister(userUpdateDurationSeconds)

}

// FindAll to satisfy api2go.FindAll interface
func (u UserResource) FindAll(r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusMethodNotAllowed
	defer userFindAllDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	return &Response{}, api2go.NewHTTPError(fmt.Errorf("Find all users not supported"), http.StatusText(code), code)
}

// FindOne to satisfy api2go.CRUD interface
func (u UserResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusMethodNotAllowed
	defer userFindAllDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	return &Response{}, api2go.NewHTTPError(fmt.Errorf("Find one users not supported"), http.StatusText(code), code)
}

// Create to satisfy api2go.CRUD interface
func (u UserResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer userCreateDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	// verify that the user is authenticated and extract firebaseUID
	requestingUserFirebaseUID, err := verify(r, u.TokenVerifier)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("error creating user: %s", err),
			http.StatusText(code),
			code,
		)
	}

	// check if the user already has a user (they might not if they are creating for themselves)
	requestingUser, err := u.UserStorage.GetByFirebaseUID(requestingUserFirebaseUID, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		// user doesn't exist yet. Create the requesting user so we can reason over if as if it already exists
		requestingUser = model.User{FirebaseUID: requestingUserFirebaseUID}
		break
	default:
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Error creating user: error occurred while retrieving requesting user, %s", err),
			http.StatusText(code),
			code,
		)
	}

	user, ok := obj.(model.User)
	if !ok {
		code = http.StatusBadRequest
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to user create: %v", obj),
			http.StatusText(code),
			code,
		)
	}

	msg, status, err := u.validateUpsert(user, requestingUser, r.Context)
	if err != nil {
		code = status
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("error creating user, %s", err),
			msg,
			code,
		)
	}

	id, err := u.UserStorage.Insert(user, r.Context)
	if err == nil && id == "" {
		err = errors.New("null id returned")
	}
	if err != nil {
		errMsg := "Error occurred while persisting user"
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			code,
		)
	}
	user.SetID(id)

	code = http.StatusCreated
	return &Response{Res: user, Code: code}, nil
}

// Delete to satisfy api2go.CRUD interface
func (u UserResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer userDeleteDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, u.TokenVerifier, u.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("error deleting user, %s", err), http.StatusText(code), code)
	}

	targetUser, err := u.UserStorage.GetOne(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusBadRequest
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("error deleting user, target user %s not found", id),
			fmt.Sprintf("error retrieving target user, %s", err),
			code,
		)
	default:
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("error deleting user, error retrieving target user %s, %s", id, err),
			fmt.Sprintf("error retrieving target user, %s", err),
			code,
		)
	}

	// not allowing firebase users to be deleted. This might be added in future, but we would need to cascade the delete so it won't be a straight forward operation
	if targetUser.FirebaseUID != "" {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("error deleting user, user %s attempting to delete firebase user %s", requestingUser.GetID(), targetUser.GetID()),
			"unable to delete users linked to Firebase",
			code,
		)
	}

	carShare, err := u.CarShareStorage.GetOne(targetUser.LinkedCarShareID, r.Context)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("error deleting user, error retrieving target user %s linked car share %s, %s", targetUser.GetID(), targetUser.LinkedCarShareID, err),
			fmt.Sprintf("error retrieving target user linked car share, %s", err),
			code,
		)
	}

	if !isAdmin(requestingUser, carShare) {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("error deleting user, user %s attempting to delete user %s linked to car share %s, but isn't an admin", requestingUser.GetID(), targetUser.GetID(), targetUser.LinkedCarShareID),
			fmt.Sprintf("only admins for car share %s can delete user %s", targetUser.LinkedCarShareID, targetUser.GetID()),
			code,
		)
	}

	for i, memberID := range carShare.MemberIDs {
		if memberID == targetUser.GetID() {
			carShare.MemberIDs = append(carShare.MemberIDs[:i], carShare.MemberIDs[i+1:]...)
			err = u.CarShareStorage.Update(carShare, r.Context)
			if err != nil {
				code = http.StatusInternalServerError
				return &Response{}, api2go.NewHTTPError(
					fmt.Errorf("error deleting user, error removing user %s from carshare %s member list, %v", targetUser.GetID(), carShare.GetID(), err),
					fmt.Sprintf("error removing user %s from carshare %s member list, %v", targetUser.GetID(), carShare.GetID(), err),
					code,
				)
			}
			break
		}
	}

	err = u.UserStorage.Delete(id, r.Context)
	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("unable to find trip %s to user", id), http.StatusText(code), code)
	default:
		errMsg := fmt.Sprintf("Error occurred while deleting user %s", id)
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	code = http.StatusOK
	return &Response{Code: code}, err
}

// Update to satisfy api2go.CRUD interface
func (u UserResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	// metrics collection. Need to be careful to capture return code before returning
	start := time.Now()
	code := http.StatusInternalServerError
	defer userUpdateDurationSeconds.WithLabelValues(fmt.Sprintf("%d", code)).Observe(time.Since(start).Seconds())

	requestingUser, err := getRequestUser(r, u.TokenVerifier, u.UserStorage)
	if err != nil {
		code = http.StatusForbidden
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Error updating user, %s", err), http.StatusText(code), code)
	}

	user, ok := obj.(model.User)
	if !ok {
		code = http.StatusBadRequest
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip update: %v", obj),
			http.StatusText(code),
			code,
		)
	}

	msg, status, err := u.validateUpsert(user, requestingUser, r.Context)
	if err != nil {
		code = status
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("Error updating user, %s", err), msg, code)
	}

	switch u.UserStorage.Update(user, r.Context) {
	case nil:
		break
	case storage.ErrNotFound:
		code = http.StatusNotFound
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Unable to find user %s to update", user.GetID()),
			http.StatusText(code),
			code,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while updating user %s", user.GetID())
		code = http.StatusInternalServerError
		return &Response{}, api2go.NewHTTPError(fmt.Errorf("%s, %s", errMsg, err), errMsg, code)
	}

	code = http.StatusNoContent
	return &Response{Res: user, Code: code}, err
}

func (u UserResource) validateUpsert(user model.User, requestingUser model.User, context api2go.APIContexter) (msg string, status int, err error) {

	if user.FirebaseUID == "" && user.LinkedCarShareID == "" {
		return "user not associated with a FirebaseUID or a LinkedCarShareID",
			http.StatusBadRequest,
			fmt.Errorf("user not associated with a FirebaseUID or a LinkedCarShareID")
	}

	if user.FirebaseUID != "" && user.FirebaseUID != requestingUser.FirebaseUID {
		return "cannot create/update a user associated with another firebase user",
			http.StatusForbidden,
			fmt.Errorf("user %s (firebaseUID %s) attempting to create/update user %s (firebaseUID %s)", requestingUser.GetID(), requestingUser.FirebaseUID, user.GetID(), user.FirebaseUID)
	}

	if user.LinkedCarShareID != "" {

		linkedCarShare, err := u.CarShareStorage.GetOne(user.LinkedCarShareID, context)
		switch err {
		case nil:
			break
		case storage.ErrInvalidID:
			return "invalid linked car share id",
				http.StatusBadRequest,
				fmt.Errorf("user %s attempting to create/update carshare linked user with invalid linked car share id \"%s\"", requestingUser.GetID(), user.LinkedCarShareID)
		case storage.ErrNotFound:
			return "linked car share not found",
				http.StatusBadRequest,
				fmt.Errorf("user %s attempting to create/update carshare linked user with non-existant linked car share id \"%s\"", requestingUser.GetID(), user.LinkedCarShareID)
		default:
			return "error finding linked car share",
				http.StatusInternalServerError,
				fmt.Errorf("error finding linked car share %v", err)
		}

		if !isAdmin(requestingUser, linkedCarShare) {
			return "requesting user not admin for carshare that target user is linked to",
				http.StatusInternalServerError,
				fmt.Errorf("user %s creating/editing user linked to car share %s which they are not admin for", requestingUser.GetID(), user.LinkedCarShareID)
		}

	}

	return "", 0, nil
}

func isAdmin(user model.User, carShare model.CarShare) bool {
	userIsAdmin := false
	for _, adminUID := range carShare.AdminIDs {
		if adminUID == user.GetID() {
			userIsAdmin = true
			break
		}
	}
	return userIsAdmin
}
