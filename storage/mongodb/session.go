package mongodb

import (
	"errors"

	"github.com/manyminds/api2go"
	mgo "gopkg.in/mgo.v2"
)

// ErrorNoDBSessionInContext request context is missing database session
var ErrorNoDBSessionInContext = errors.New("Error retrieving mongodb session from context")

// ErrorInvalidDBSession unable to case database session in request context as mgo.Session
var ErrorInvalidDBSession = errors.New("Error asserting type of mongodb session from context")

func getMgoSession(context api2go.APIContexter) (*mgo.Session, error) {
	ctxMgoSession, ok := context.Get("db")
	if !ok {
		return nil, ErrorNoDBSessionInContext
	}

	mgoSession, ok := ctxMgoSession.(*mgo.Session)
	if !ok {
		return nil, ErrorInvalidDBSession
	}

	return mgoSession.Clone(), nil
}
