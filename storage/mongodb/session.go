package mongodb_storage

import (
	"errors"

	"github.com/manyminds/api2go"
	mgo "gopkg.in/mgo.v2"
)

func getMgoSession(context api2go.APIContexter) (*mgo.Session, error) {
	ctxMgoSession, ok := context.Get("db")
	if !ok {
		return nil, errors.New("Error retrieving mongodb session from context")
	}

	mgoSession, ok := ctxMgoSession.(*mgo.Session)
	if !ok {
		return nil, errors.New("Error asserting type of mongodb session from context")
	}

	return mgoSession.Clone(), nil
}
