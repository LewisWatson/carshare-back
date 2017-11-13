package resource

import (
	"gopkg.in/jose.v1/jwt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	dockertest "gopkg.in/ory-am/dockertest.v3"
	mgo "gopkg.in/mgo.v2"

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

type mockTokenVerifier struct {
	Claims jwt.Claims
	Error  error
}

func (mtv mockTokenVerifier) Verify(accessToken string) (userID string, claims jwt.Claims, err error) {
	return mtv.Claims.Get("sub").(string), mtv.Claims, mtv.Error
}
