package mongodb-storage

import (
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// CarShareStorage stores all car shares
type CarShareStorage struct{}

// GetAll to satisfy storage.CarShareStoreage interface
func (s CarShareStorage) GetAll(context api2go.APIContexter) ([]model.CarShare, error) {
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return nil, err
	}
	defer mgoSession.Close()
	result := []model.CarShare{}
	err = mgoSession.DB("carshare").C("carShares").Find(nil).All(&result)
	if err != nil {
		log.Printf("Error finding all car shares, %s", err)
	}
	return result, err
}

// GetOne to satisfy storage.CarShareStoreage interface
func (s CarShareStorage) GetOne(id string, context api2go.APIContexter) (model.CarShare, error) {
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return model.CarShare{}, err
	}
	defer mgoSession.Close()
	result := model.CarShare{}
	err = mgoSession.DB("carshare").C("carShares").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	if err != nil {
		log.Printf("Error finding car share %s, %s", id, err)
		if err == mgo.ErrNotFound {
			err = storage.ErrNotFound
		}
	}
	return result, err
}

// Insert to satisfy storage.CarShareStoreage interface
func (s *CarShareStorage) Insert(c model.CarShare, context api2go.APIContexter) (string, error) {
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return "", err
	}
	defer mgoSession.Close()
	c.ID = bson.NewObjectId()
	err = mgoSession.DB("carshare").C("carShares").Insert(&c)
	if err != nil {
		log.Printf("Error inserting car share, %s", err)
	}
	return c.GetID(), err
}

// Delete to satisfy storage.CarShareStoreage interface
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
	if err != nil {
		log.Printf("Error finding all car shares, %s", err)
		if err == mgo.ErrNotFound {
			err = storage.ErrNotFound
		}
	}
	return err
}

// Update to satisfy storage.CarShareStoreage interface
func (s *CarShareStorage) Update(c model.CarShare, context api2go.APIContexter) error {
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()
	err = mgoSession.DB("carshare").C("carShares").Update(bson.M{"_id": c.ID}, &c)
	if err != nil {
		log.Printf("Error updating car share, %s", err)
		if err == mgo.ErrNotFound {
			err = storage.ErrNotFound
		}
	}
	return err
}
