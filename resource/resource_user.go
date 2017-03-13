package resource

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
	"gopkg.in/LewisWatson/firebase-jwt-auth.v1"
)

// UserResource for api2go routes
type UserResource struct {
	UserStorage     storage.UserStorage
	CarShareStorage storage.CarShareStorage
	TokenVerifier   fireauth.TokenVerifier
}

// FindAll to satisfy api2go.FindAll interface
func (u UserResource) FindAll(r api2go.Request) (api2go.Responder, error) {
	return &Response{}, api2go.NewHTTPError(
		fmt.Errorf("Find all users not supported"),
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
}

// FindOne to satisfy api2go.CRUD interface
func (u UserResource) FindOne(ID string, r api2go.Request) (api2go.Responder, error) {
	return &Response{}, api2go.NewHTTPError(
		fmt.Errorf("Fine one users not supported"),
		http.StatusText(http.StatusMethodNotAllowed),
		http.StatusMethodNotAllowed,
	)
}

// Create to satisfy api2go.CRUD interface
func (u UserResource) Create(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	userID, err := verify(r, u.TokenVerifier)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Error creating user, %s", err),
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden,
		)
	}

	user, ok := obj.(model.User)
	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to user create: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}

	if user.FirebaseUID == "" && user.LinkedCarShareID == "" {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("A user must be associated with either a FirebaseUID or a LinkedCarShareID"),
			"A user must be associated with either a FirebaseUID or a LinkedCarShareID",
			http.StatusBadRequest,
		)
	}

	if user.FirebaseUID != "" && user.LinkedCarShareID != "" {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("A user can have a FirebaseUID or a LinkedCarShareID, not both. FirebaseUID %s LinkedCarShareID %s", user.FirebaseUID, user.LinkedCarShareID),
			"A user can have a FirebaseUID or a LinkedCarShareID, not both",
			http.StatusBadRequest,
		)
	}

	if user.FirebaseUID != "" && user.FirebaseUID != userID {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("FirebaseUID \"%s\" attempting to create user with FirebaseUID \"%s\"", userID, user.FirebaseUID),
			"Cannot create a user linked to another firebase user",
			http.StatusBadRequest,
		)
	}

	if user.LinkedCarShareID != "" {

		_, err = u.CarShareStorage.GetOne(user.LinkedCarShareID, r.Context)

		switch err {
		case nil:
			break
		case storage.ErrInvalidID:
		case storage.ErrNotFound:
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("User %s attempting to create a user linked to a non existant car share %s %v", userID, user.LinkedCarShareID, err),
				fmt.Sprintf("unable to find car share %s", user.LinkedCarShareID),
				http.StatusBadRequest,
			)
		default:
			return &Response{}, api2go.NewHTTPError(
				fmt.Errorf("Error retrieving linked car share when creating user %+v, %s", user, err),
				http.StatusText(http.StatusInternalServerError),
				http.StatusInternalServerError,
			)
		}

	}

	id, err := u.UserStorage.Insert(user, r.Context)
	if err == nil && id == "" {
		err = errors.New("null id returned")
	}
	if err != nil {
		errMsg := "Error occurred while persisting user"
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}
	user.SetID(id)

	return &Response{Res: user, Code: http.StatusCreated}, nil
}

// Delete to satisfy api2go.CRUD interface
func (u UserResource) Delete(id string, r api2go.Request) (api2go.Responder, error) {

	err := u.UserStorage.Delete(id, r.Context)

	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("unable to find trip %s to user", id),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while deleting user %s", id)
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	return &Response{Code: http.StatusOK}, err
}

// Update to satisfy api2go.CRUD interface
func (u UserResource) Update(obj interface{}, r api2go.Request) (api2go.Responder, error) {

	firebaseUID, err := verify(r, u.TokenVerifier)
	if err != nil {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Error updating user, %s", err),
			http.StatusText(http.StatusForbidden),
			http.StatusForbidden,
		)
	}

	user, ok := obj.(model.User)
	if !ok {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Invalid instance given to trip update: %v", obj),
			http.StatusText(http.StatusBadRequest),
			http.StatusBadRequest,
		)
	}

	if user.FirebaseUID != "" && user.FirebaseUID != firebaseUID {
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("FirebaseUID \"%s\" attempting to update user with FirebaseUID \"%s\"", firebaseUID, user.FirebaseUID),
			"cannot update a user for another firebase user",
			http.StatusForbidden,
		)
	}

	err = u.UserStorage.Update(user, r.Context)

	switch err {
	case nil:
		break
	case storage.ErrNotFound:
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("Unable to find user %s to update", user.GetID()),
			http.StatusText(http.StatusNotFound),
			http.StatusNotFound,
		)
	default:
		errMsg := fmt.Sprintf("Error occurred while updating user %s", user.GetID())
		return &Response{}, api2go.NewHTTPError(
			fmt.Errorf("%s, %s", errMsg, err),
			errMsg,
			http.StatusInternalServerError,
		)
	}

	return &Response{Res: user, Code: http.StatusNoContent}, err
}
