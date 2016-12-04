package mongodb_storage

import (
	"sort"

	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// sorting
type byID []model.Trip

func (t byID) Len() int {
	return len(t)
}

func (t byID) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t byID) Less(i, j int) bool {
	return t[i].GetID() < t[j].GetID()
}

type TripStorage struct{}

// GetAll of the trips
func (s TripStorage) GetAll(context api2go.APIContexter) ([]model.Trip, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return nil, err
	}
	defer mgoSession.Close()

	result := []model.Trip{}
	err = mgoSession.DB("carshare").C("trips").Find(nil).All(&result)
	sort.Sort(byID(result))
	s.setTimezonesToUTC(&result)
	return result, err
}

// GetOne trip
func (s TripStorage) GetOne(id string, context api2go.APIContexter) (model.Trip, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return model.Trip{}, err
	}
	defer mgoSession.Close()

	result := model.Trip{}
	err = mgoSession.DB("carshare").C("trips").Find(bson.M{"_id": bson.ObjectIdHex(id)}).One(&result)
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	s.setTimezoneToUTC(&result)
	return result, err
}

// Insert a fresh one
func (s *TripStorage) Insert(t model.Trip, context api2go.APIContexter) (string, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return "", err
	}
	defer mgoSession.Close()

	t.ID = bson.NewObjectId()
	err = mgoSession.DB("carshare").C("trips").Insert(&t)
	if err != nil {
		return "", err
	}
	return t.GetID(), nil
}

// Delete one :(
func (s *TripStorage) Delete(id string, context api2go.APIContexter) error {

	if !bson.IsObjectIdHex(id) {
		return storage.InvalidID
	}

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()

	err = mgoSession.DB("carshare").C("trips").Remove(bson.M{"_id": bson.ObjectIdHex(id)})
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	return err
}

// Update updates an existing trip
func (s *TripStorage) Update(t model.Trip, context api2go.APIContexter) error {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return err
	}
	defer mgoSession.Close()

	err = mgoSession.DB("carshare").C("trips").Update(bson.M{"_id": t.ID}, &t)
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	return err
}

func (s *TripStorage) GetLatest(carShareID string, context api2go.APIContexter) (model.Trip, error) {

	mgoSession, err := getMgoSession(context)
	if err != nil {
		return model.Trip{}, err
	}
	defer mgoSession.Close()

	latestTrip := model.Trip{}
	err = mgoSession.DB("carshare").C("trips").Find(bson.M{"car-share": carShareID}).Sort("-timestamp").One(&latestTrip)
	if err == mgo.ErrNotFound {
		err = storage.ErrNotFound
	}
	s.setTimezoneToUTC(&latestTrip)
	return latestTrip, err
}

func (s *TripStorage) setTimezonesToUTC(trips *[]model.Trip) {
	for _, trip := range *trips {
		s.setTimezoneToUTC(&trip)
	}
}

// time.Time values get stored in MongoDB as timestamps without timezones
// when they are read they are given a timezone, we want to ensure we stick
// to UTC at all times
func (s *TripStorage) setTimezoneToUTC(trip *model.Trip) {
	trip.TimeStamp = trip.TimeStamp.UTC()
}
