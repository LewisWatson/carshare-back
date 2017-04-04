package mongodb

import (
	"fmt"

	mgo "gopkg.in/mgo.v2"
	dockertest "gopkg.in/ory-am/dockertest.v3"
)

// ConnectToMongoDB spin up a mongodb for integration tests
func ConnectToMongoDB(db *mgo.Session, pool *dockertest.Pool, containerResource *dockertest.Resource) (*mgo.Session, *dockertest.Pool, *dockertest.Resource) {

	if db == nil {

		containerName := "mongo"
		version := "3.4"

		log.Infof("Spinning up %s:%s container\n", containerName, version)

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

		log.Info("Connection to MongoDB established")
	}

	return db, pool, containerResource
}
