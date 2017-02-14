package resource

import (
	"github.com/LewisWatson/carshare-back/auth"
	"github.com/manyminds/api2go"
)

func verify(r api2go.Request, tokenVerifier auth.TokenVerifier) error {
	token := r.Header.Get("authorization")
	claims, err := tokenVerifier.Verify(token)
	if err != nil {
		return err
	}
	r.Context.Set("userID", claims.Get("user_id"))
	return nil
}
