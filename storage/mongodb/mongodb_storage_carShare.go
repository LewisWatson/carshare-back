package mongodb_storage

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// NewCarShareStorage initializes the storage
func NewCarShareStorage(db *mgo.Session) *CarShareStorage {
	return &CarShareStorage{db.DB("carshare").C("carShares")}
}

// CarShareStorage stores all car shares
type CarShareStorage struct {
	carShares *mgo.Collection
}

func (s CarShareStorage) GetAll(context api2go.APIContexter) ([]model.CarShare, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return nil, err
	}
	defer mgoSession.Close()

	result := []model.CarShare{}
	err = mgoSession.DB("carshare").C("carShares").Find(nil).All(&result)
	return result, err
}

// GetOne carShare
func (s CarShareStorage) GetOne(id string, context api2go.APIContexter) (model.CarShare, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return model.CarShare{}, err
	}
	defer mgoSession.Close()

	result := model.CarShare{}
	err = mgoSession.DB("carshare").C("carShares").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	return result, err
}

// Insert a carShare
func (s *CarShareStorage) Insert(c model.CarShare, context api2go.APIContexter) (string, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return "", err
	}
	defer mgoSession.Close()

	c.ID = bson.NewObjectId()
	err = mgoSession.DB("carshare").C("carShares").Insert(&c)
	return c.GetID(), err
}

// Delete one :(
func (s *CarShareStorage) Delete(id string, context api2go.APIContexter) error {

	if !bson.IsObjectIdHex(id) {
		return storage.InvalidID
	}

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()

	err = mgoSession.DB("carshare").C("carShares").Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	return err
}

// Update a carShare
func (s *CarShareStorage) Update(c model.CarShare, context api2go.APIContexter) error {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()

	err = mgoSession.DB("carshare").C("carShares").Update(bson.M{"_id": c.ID}, &c)
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	return err
}
