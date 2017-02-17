package auth

import "github.com/SermoDigital/jose/jwt"

// TokenVerifier verifies authenticaion tokens
type TokenVerifier interface {
	Verify(token string) (userID string, claims jwt.Claims, err error)
}
