package resource

import (
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
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