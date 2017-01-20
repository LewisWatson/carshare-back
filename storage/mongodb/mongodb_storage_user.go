package mongodb-storage

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// UserStorage stores all users
type UserStorage struct{}

// GetAll to satisfy storage.UserStorage interface
func (s UserStorage) GetAll(context api2go.APIContexter) ([]model.User, error) {
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return nil, err
	}
	defer mgoSession.Close()
	result := []model.User{}
	err = mgoSession.DB("carshare").C("users").Find(nil).All(&result)
	return result, err
}

// GetOne to satisfy storage.UserStorage interface
func (s UserStorage) GetOne(id string, context api2go.APIContexter) (model.User, error) {
	if !bson.IsObjectIdHex(id) {
		return model.User{}, storage.InvalidID
	}
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return model.User{}, err
	}
	defer mgoSession.Close()
	result := model.User{}
	err = mgoSession.DB("carshare").C("users").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	return result, err
}

// Insert to satisfy storage.UserStorage interface
func (s *UserStorage) Insert(u model.User, context api2go.APIContexter) (string, error) {
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return "", err
	}
	defer mgoSession.Close()
	u.ID = bson.NewObjectId()
	err = mgoSession.DB("carshare").C("users").Insert(&u)
	return u.GetID(), err
}

// Delete to satisfy storage.UserStorage interface
func (s *UserStorage) Delete(id string, context api2go.APIContexter) error {
	if !bson.IsObjectIdHex(id) {
		return storage.InvalidID
	}
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()
	err = mgoSession.DB("carshare").C("users").Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	return err
}

// Update to satisfy storage.UserStorage interface
func (s *UserStorage) Update(u model.User, context api2go.APIContexter) error {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()

	err = mgoSession.DB("carshare").C("users").Update(bson.M{"_id": u.ID}, &u)
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	return err
}
