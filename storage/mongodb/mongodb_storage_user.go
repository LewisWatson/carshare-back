package mongodb_storage

import (
	"errors"
	"fmt"
	"net/http"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// UserStorage stores all users
type UserStorage struct{}

// GetAll of the users
func (s UserStorage) GetAll(context api2go.APIContexter) ([]model.User, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return nil, err
	}
	defer mgoSession.Close()

	result := []model.User{}
	err = getUsersCollection(mgoSession).Find(nil).All(&result)
	if err != nil {
		return nil, err
	}
	return result, nil
}

// GetOne user
func (s UserStorage) GetOne(id string, context api2go.APIContexter) (model.User, error) {

	if !bson.IsObjectIdHex(id) {
		return model.User{}, errors.New(fmt.Sprintf("Error retrieving user %s, Invalid ID", id))
	}

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return model.User{}, err
	}
	defer mgoSession.Close()

	result := model.User{}
	err = getUsersCollection(mgoSession).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	if err != nil {
		errMessage := fmt.Sprintf("Error retrieving user %s, %s", id, err)
		return model.User{}, api2go.NewHTTPError(errors.New(errMessage), errMessage, http.StatusNotFound)
	}
	return result, nil
}

// Insert a user
func (s *UserStorage) Insert(u model.User, context api2go.APIContexter) (string, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return "", err
	}
	defer mgoSession.Close()

	u.ID = bson.NewObjectId()
	err = getUsersCollection(mgoSession).Insert(&u)
	if err != nil {
		return "", errors.New(fmt.Sprintf("Error inserting user %s, %s", u.GetID(), err))
	}
	return u.GetID(), nil
}

// Delete one :(
func (s *UserStorage) Delete(id string, context api2go.APIContexter) error {

	if !bson.IsObjectIdHex(id) {
		return errors.New(fmt.Sprintf("Error deleting user %s, Invalid ID", id))
	}

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()

	err = getUsersCollection(mgoSession).Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil {
		return errors.New(fmt.Sprintf("Error deleting user %s, %s", id, err))
	}
	return nil
}

// Update a user
func (s *UserStorage) Update(u model.User, context api2go.APIContexter) error {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()

	err = getUsersCollection(mgoSession).Update(bson.M{"_id": u.ID}, &u)
	if err != nil {
		return errors.New(fmt.Sprintf("Error updating user %s, %s", u.GetID(), err))
	}
	return nil
}

func getUsersCollection(mgoSession *mgo.Session) *mgo.Collection {
	return mgoSession.DB("carshare").C("users")
}

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
