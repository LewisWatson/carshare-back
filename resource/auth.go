package resource

import (
	"fmt"

	"github.com/LewisWatson/carshare-back/storage"
	"github.com/LewisWatson/firebase-jwt-auth"
	"github.com/manyminds/api2go"
)

func verify(r api2go.Request, tokenVerifier fireauth.TokenVerifier, userStorage storage.UserStorage) (string, error) {
	token := r.Header.Get("authorization")
	userID, claims, err := tokenVerifier.Verify(token)
	if err != nil {
		return "", err
	}

	_, err = userStorage.GetOne(userID, r.Context)
	if err != nil {
		return "", fmt.Errorf("Unable to retrieve user specified in auth token, " + err.Error())
	}

	r.Context.Set("userID", userID)
	r.Context.Set("claims", claims)
	return userID, nil
}
