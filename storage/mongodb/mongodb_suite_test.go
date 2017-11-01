package mongodb

import (
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	mgo "gopkg.in/mgo.v2"
	dockertest "github.com/ory-am/dockertest"

	"testing"
)

var (
	db                *mgo.Session
	pool              *dockertest.Pool
	containerResource *dockertest.Resource
)

func TestMongodb(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Mongodb Suite")
}

var _ = AfterSuite(func() {

	if db != nil {
		log.Info("Closing connection to MongoDB")
		db.Close()
	}

	if pool != nil {
		log.Info("Purging containers")
		if err := pool.Purge(containerResource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
})
