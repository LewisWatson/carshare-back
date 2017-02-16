package auth

import "errors"

var (
	// ErrNilToken is returned when the authorization token is empty
	ErrNilToken = errors.New("Empty authorizatin token")

	// ErrTokenExpired is returnd when time.Now().Unix() is after the token's "exp" claim
	ErrTokenExpired = errors.New("Token is expired")

	// ErrECDSAVerification is missing from crypto/ecdsa compared to crypto/rsa
	ErrECDSAVerification = errors.New("crypto/ecdsa: verification error")

	// ErrNotCompact signals that the provided potential JWS is not in its compact representation.
	ErrNotCompact = errors.New("not a compact JWS")

	// ErrInvalidAud indicates that the authorisation token audience claim is invalid
	ErrInvalidAud = errors.New("Invalid auth token audience")

	// ErrInvalidIss indicates that the authorisation token issuer is invalid
	ErrInvalidIss = errors.New("Invalid auth token issuer")
)
