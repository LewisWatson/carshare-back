package resource

import (
	"errors"
	"fmt"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/LewisWatson/firebase-jwt-auth"
	"github.com/manyminds/api2go"
)

// verify the request auth token
func verify(r api2go.Request, tokenVerifier fireauth.TokenVerifier) (firebaseUID string, err error) {
	token := r.Header.Get("authorization")
	userID, claims, err := tokenVerifier.Verify(token)
	if err != nil {
		return "", err
	}
	r.Context.Set("userID", userID)
	r.Context.Set("claims", claims)
	return userID, nil
}

// getRequestUser verifies the request authorisation token and finds the user it links to
func getRequestUser(r api2go.Request, tokenVerifier fireauth.TokenVerifier, userStorage storage.UserStorage) (requestUser model.User, err error) {
	firebaseUID, err := verify(r, tokenVerifier)
	if err != nil {
		return model.User{}, err
	}
	requestUser, err = userStorage.GetByFirebaseUID(firebaseUID, r.Context)
	if err == storage.ErrNotFound {
		requestUser, err = createAppUserForFirebaseUser(firebaseUID, r, userStorage)
	}
	return requestUser, err
}

// createAppUserForFirebaseUser inserts a new user into user storage with the provided firebaseUID
func createAppUserForFirebaseUser(firebaseUID string, r api2go.Request, userStorage storage.UserStorage) (user model.User, err error) {
	user = model.User{FirebaseUID: firebaseUID}
	var id string
	id, err = userStorage.Insert(user, r.Context)
	if err == nil && id == "" {
		err = errors.New("null id returned")
	}
	if err != nil {
		err = fmt.Errorf("error creating new user, %s", err)
	} else {
		user.SetID(id)
	}
	return user, err
}
