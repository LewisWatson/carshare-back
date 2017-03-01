package storage

import "errors"

// ErrNotFound indicates that an entity has not been found
var ErrNotFound = errors.New("not found")

// ErrInvalidID indicates that the provided ID is not valid
var ErrInvalidID = errors.New("invalid ID")
