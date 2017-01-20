package mongodb-storage

import (
	"log"
	"sort"

	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
)

// TripStorage a place to store car share trips
type TripStorage struct {
	CarshareStorage CarShareStorage
}

// GetAll to satisfy storage.TripStoreage interface
func (s TripStorage) GetAll(carShareID string, context api2go.APIContexter) ([]model.Trip, error) {
	carShare, err := s.CarshareStorage.GetOne(carShareID, context)
	if err != nil {
		return nil, err
	}
	return carShare.Trips, nil
}

// GetOne to satisfy storage.TripStoreage interface
func (s TripStorage) GetOne(carShareID string, id string, context api2go.APIContexter) (model.Trip, error) {
	if !bson.IsObjectIdHex(id) {
		return model.Trip{}, storage.InvalidID
	}
	carShare, err := s.CarshareStorage.GetOne(carShareID, context)
	if err != nil {
		return model.Trip{}, err
	}
	trip, err := s.findTrip(id, carShare)
	if err != nil {
		log.Printf("Erorr finding trip %s in car share %s, %s", id, carShare.GetID(), err)
		return model.Trip{}, err
	}
	s.setTimezoneToUTC(&trip)
	return trip, nil
}

// Insert to satisfy storage.TripStoreage interface
func (s *TripStorage) Insert(carShareID string, t model.Trip, context api2go.APIContexter) (string, error) {
	carShare, err := s.CarshareStorage.GetOne(carShareID, context)
	if err != nil {
		return "", err
	}
	t.ID = bson.NewObjectId()
	carShare.Trips = append(carShare.Trips, t)
	sort.Sort(ByTimeStamp(carShare.Trips))
	err = s.CarshareStorage.Update(carShare, context)
	if err != nil {
		log.Printf("Error updating car share %s with trip, %s", t.CarShareID, err)
		return "", err
	}
	return t.GetID(), nil
}

// Delete to satisfy storage.TripStoreage interface
func (s *TripStorage) Delete(carShareID string, id string, context api2go.APIContexter) error {
	if !bson.IsObjectIdHex(id) {
		return storage.InvalidID
	}
	carShare, err := s.CarshareStorage.GetOne(carShareID, context)
	if err != nil {
		return err
	}
	for index, trip := range carShare.Trips {
		if trip.GetID() == id {
			carShare.Trips = append(carShare.Trips[:index], carShare.Trips[index+1:]...)
			break
		}
	}
	sort.Sort(ByTimeStamp(carShare.Trips))
	return s.CarshareStorage.Update(carShare, context)
}

// Update to satisfy storage.TripStoreage interface
func (s *TripStorage) Update(carShareID string, t model.Trip, context api2go.APIContexter) error {
	carShare, err := s.CarshareStorage.GetOne(carShareID, context)
	if err != nil {
		return err
	}
	for index, trip := range carShare.Trips {
		if trip.GetID() == t.GetID() {
			carShare.Trips[index] = trip
			break
		}
	}
	sort.Sort(ByTimeStamp(carShare.Trips))
	return s.CarshareStorage.Update(carShare, context)
}

// GetLatest to satisfy storage.TripStoreage interface
func (s *TripStorage) GetLatest(carShareID string, context api2go.APIContexter) (model.Trip, error) {
	carShare, err := s.CarshareStorage.GetOne(carShareID, context)
	if err != nil {
		log.Printf("Error finding car share %s, %s", carShareID, err)
		return model.Trip{}, err
	}
	if carShare.Trips == nil {
		return model.Trip{}, storage.ErrNotFound
	}
	sort.Sort(ByTimeStamp(carShare.Trips))
	return carShare.Trips[0], nil
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

// findCarShareWithTrip finds a carshare entry with a trip subdocument with a matching id
func (s TripStorage) findCarShareWithTrip(id string, context api2go.APIContexter) (model.CarShare, error) {
	mgoSession, err := getMgoSession(context)
	if err != nil {
		return model.CarShare{}, err
	}
	defer mgoSession.Close()
	carShare := model.CarShare{}
	err = mgoSession.DB("carshare").C("carshares").Find(bson.M{"trips._id": bson.ObjectIdHex(id)}).One(&carShare)
	return carShare, err
}

// findTrip finds a trip with a matching id from a car shares list of trips
func (s TripStorage) findTrip(id string, carShare model.CarShare) (model.Trip, error) {
	for index, trip := range carShare.Trips {
		if trip.GetID() == id {
			return carShare.Trips[index], nil
		}
	}
	return model.Trip{}, storage.ErrNotFound
}

// ByTimeStamp implements sort.Interface for []model.Trip based on the TimeStamp field.
type ByTimeStamp []model.Trip

// Len return length of array
func (a ByTimeStamp) Len() int { return len(a) }

// Swap swap items in sli
func (a ByTimeStamp) Swap(i, j int) { a[i], a[j] = a[j], a[i] }

// Less return true if trip i occured before j
func (a ByTimeStamp) Less(i, j int) bool { return a[i].TimeStamp.Before(a[j].TimeStamp) }
