package mongodb_test

import (
	"fmt"
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
	RunSpecs(t, "Mongodb Suite")
}

var connectToMongoDB = func() {

	if db != nil {
		return
	}

	containerName := "mongo"
	version := "3.4"

	fmt.Println()
	log.Printf("Spinning up %s:%s container\n", containerName, version)

	var err error

	pool, err = dockertest.NewPool("")
	if err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	containerResource, err = pool.Run(containerName, version, []string{"--smallfiles"})
	if err != nil {
		log.Fatalf("Could not start resource: %s", err)
	}

	if err = pool.Retry(func() error {
		db, err = mgo.Dial(fmt.Sprintf("localhost:%s", containerResource.GetPort("27017/tcp")))
		if err != nil {
			return err
		}
		return db.Ping()
	}); err != nil {
		log.Fatalf("Could not connect to docker: %s", err)
	}

	log.Println("Connection to MongoDB established")
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
