package resource

import (
	"fmt"

	"github.com/LewisWatson/carshare-back/storage"
	"github.com/LewisWatson/firebase-jwt-auth"
	"github.com/manyminds/api2go"
)

// verify the request auth token
func verify(r api2go.Request, tokenVerifier fireauth.TokenVerifier) (string, error) {
	token := r.Header.Get("authorization")
	userID, claims, err := tokenVerifier.Verify(token)
	if err != nil {
		return "", err
	}
	r.Context.Set("userID", userID)
	r.Context.Set("claims", claims)
	return userID, nil
}

// verify the request auth token is valid and refers to an existin user
func verifyUserExists(r api2go.Request, tokenVerifier fireauth.TokenVerifier, userStorage storage.UserStorage) (string, error) {
	userID, err := verify(r, tokenVerifier)
	if err == nil {
		_, err = userStorage.GetOne(userID, r.Context)
		if err != nil {
			return "", fmt.Errorf("Unable to retrieve user specified in auth token, " + err.Error())
		}
	}
	return userID, err
}
