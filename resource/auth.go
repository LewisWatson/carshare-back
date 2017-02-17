package resource

import (
	"github.com/LewisWatson/carshare-back/auth"
	"github.com/manyminds/api2go"
)

func verify(r api2go.Request, tokenVerifier auth.TokenVerifier) error {
	token := r.Header.Get("authorization")
	userID, claims, err := tokenVerifier.Verify(token)
	if err != nil {
		return err
	}
	r.Context.Set("userID", userID)
	r.Context.Set("claims", claims)
	return nil
}
