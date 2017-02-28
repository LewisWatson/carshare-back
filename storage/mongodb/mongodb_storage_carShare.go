package mongodb

import (
	"log"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// CarShareStorage stores all car shares
type CarShareStorage struct{}

// GetAll to satisfy storage.CarShareStorage interface
func (s CarShareStorage) GetAll(userID string, ctx api2go.APIContexter) ([]model.CarShare, error) {
	ms, err := getMgoSession(ctx)
	if err != nil {
		return nil, err
	}
	defer ms.Close()
	result := []model.CarShare{}
	err = ms.DB(CarShareDB).C(CarSharesColl).Find(bson.M{"members": userID}).All(&result)
	return result, err
}

// GetOne to satisfy storage.CarShareStoreage interface
func (s CarShareStorage) GetOne(id string, ctx api2go.APIContexter) (model.CarShare, error) {

	if !bson.IsObjectIdHex(id) {
		return model.CarShare{}, storage.ErrInvalidID
	}

	ms, err := getMgoSession(ctx)
	if err != nil {
		return model.CarShare{}, err
	}
	defer ms.Close()
	result := model.CarShare{}
	err = ms.DB(CarShareDB).C(CarSharesColl).Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
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
	err = mgoSession.DB(CarShareDB).C(CarSharesColl).Insert(&c)
	if err != nil {
		log.Printf("Error inserting car share, %s", err)
	}
	return c.GetID(), err
}

// Delete to satisfy storage.CarShareStoreage interface
func (s *CarShareStorage) Delete(id string, context api2go.APIContexter) error {
	if !bson.IsObjectIdHex(id) {
		return storage.ErrInvalidID
	}
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()
	err = mgoSession.DB(CarShareDB).C(CarSharesColl).Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err != nil {
		log.Printf("Error inserting car share, %s", err)
		if err == mgo.ErrNotFound {
			err = storage.ErrNotFound
		}
	}
	return err
}

// Update to satisfy storage.CarShareStoreage interface
func (s *CarShareStorage) Update(c model.CarShare, context api2go.APIContexter) error {
	if !bson.IsObjectIdHex(c.GetID()) {
		return storage.ErrInvalidID
	}
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()
	err = mgoSession.DB(CarShareDB).C(CarSharesColl).Update(bson.M{"_id": c.ID}, &c)
	if err != nil {
		log.Printf("Error updating car share, %s", err)
		if err == mgo.ErrNotFound {
			err = storage.ErrNotFound
		}
	}
	return err
}
