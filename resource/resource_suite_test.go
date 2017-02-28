package resource

import (
	"log"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/jose.v1/jwt"
	mgo "gopkg.in/mgo.v2"
	dockertest "gopkg.in/ory-am/dockertest.v3"

	"testing"
)

var (
	db                *mgo.Session
	pool              *dockertest.Pool
	containerResource *dockertest.Resource
)

func TestMongodb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Resource Suite")
}

type mockTokenVerifier struct {
	Claims jwt.Claims
	Error  error
}

func (mtv mockTokenVerifier) Verify(accessToken string) (userID string, claims jwt.Claims, err error) {
	return mtv.Claims.Get("sub").(string), mtv.Claims, mtv.Error
}

type mockCarShareStorage struct {
	CarShares []model.CarShare
	CarShare  model.CarShare
	ID        string
	Error     error
}

func (css mockCarShareStorage) GetAll(userID string, context api2go.APIContexter) ([]model.CarShare, error) {
	return css.CarShares, css.Error
}

func (css mockCarShareStorage) GetOne(id string, context api2go.APIContexter) (model.CarShare, error) {
	return css.CarShare, css.Error
}

func (css mockCarShareStorage) Insert(c model.CarShare, context api2go.APIContexter) (string, error) {
	return css.ID, css.Error
}

func (css mockCarShareStorage) Delete(id string, context api2go.APIContexter) error {
	return css.Error
}

func (css mockCarShareStorage) Update(c model.CarShare, context api2go.APIContexter) error {
	return css.Error
}

type mockUserStorage struct {
	Users []model.User
	User  model.User
	ID    string
	Error error
}

func (us mockUserStorage) GetAll(context api2go.APIContexter) ([]model.User, error) {
	return us.Users, us.Error
}

func (us mockUserStorage) GetOne(id string, context api2go.APIContexter) (model.User, error) {
	return us.User, us.Error
}

func (us mockUserStorage) Insert(c model.User, context api2go.APIContexter) (string, error) {
	return us.ID, us.Error
}

func (us mockUserStorage) Delete(id string, context api2go.APIContexter) error {
	return us.Error
}

func (us mockUserStorage) Update(c model.User, context api2go.APIContexter) error {
	return us.Error
}

type mockTripStorage struct {
	Trips []model.Trip
	Trip  model.Trip
	ID    string
	Error error
}

func (ts mockTripStorage) GetAll(context api2go.APIContexter) ([]model.Trip, error) {
	return ts.Trips, ts.Error
}

func (ts mockTripStorage) GetOne(id string, context api2go.APIContexter) (model.Trip, error) {
	return ts.Trip, ts.Error
}

func (ts mockTripStorage) Insert(c model.Trip, context api2go.APIContexter) (string, error) {
	return ts.ID, ts.Error
}

func (ts mockTripStorage) Delete(id string, context api2go.APIContexter) error {
	return ts.Error
}

func (ts mockTripStorage) Update(c model.Trip, context api2go.APIContexter) error {
	return ts.Error
}

func (ts mockTripStorage) GetLatest(carShareID string, context api2go.APIContexter) (model.Trip, error) {
	return ts.Trip, ts.Error
}

var _ = AfterSuite(func() {

	if db != nil {
		log.Println("Closing connection to MongoDB")
		db.Close()
	}

	if pool != nil {
		log.Println("Purging containers")
		if err := pool.Purge(containerResource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
})
