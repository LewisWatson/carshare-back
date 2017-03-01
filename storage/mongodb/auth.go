package mongodb

import (
	"errors"

	"github.com/manyminds/api2go"
)

func getUserID(context api2go.APIContexter) (string, error) {
	userID, ok := context.Get("userID")
	if !ok {
		return "", errors.New("no userID in context")
	}

	if userID.(string) == "" {
		return "", errors.New("empty user id")
	}

	return userID.(string), nil
}
