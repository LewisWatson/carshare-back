package mongodb_test

import (
	"fmt"
	"log"

	"github.com/LewisWatson/carshare-back/model"
	. "github.com/LewisWatson/carshare-back/storage/mongodb"
	mgo "gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	dockertest "gopkg.in/ory-am/dockertest.v3"

	"github.com/manyminds/api2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var (
	db                *mgo.Session
	pool              *dockertest.Pool
	containerResource *dockertest.Resource
)

var _ = Describe("Mongodb Data Store", func() {

	Describe("Car Share Storage", func() {

		var (
			carShareStorage *CarShareStorage
			context         *api2go.APIContext
		)

		BeforeEach(func() {
			carShareStorage = &CarShareStorage{}
			context = &api2go.APIContext{}
			connectToMongoDB()
			err := db.DB("carshare").DropDatabase()
			Expect(err).ToNot(HaveOccurred())
			context.Set("db", db)
			db.DB("carshare").C("carShares").Insert(
				&model.CarShare{
					Name: "Example Car Share 1",
				},
				&model.CarShare{
					Name: "Example Car Share 2",
				},
				&model.CarShare{
					Name: "Example Car Share 3",
				},
			)
		})

		Describe("get all", func() {

			var (
				existingCarShares []model.CarShare
				result            []model.CarShare
				err               error
			)

			BeforeEach(func() {
				err = db.DB("carshare").C("carShares").Find(nil).All(&existingCarShares)
				Expect(err).ToNot(HaveOccurred())
				result, err = carShareStorage.GetAll(context)
			})

			Context("with valid mgo connection", func() {

				Context("targeting a car share that exists", func() {

				})

				It("should return all existing car shares", func() {
					Expect(err).ToNot(HaveOccurred())
					Expect(result).To(ConsistOf(existingCarShares))
				})

			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					result, err = carShareStorage.GetAll(context)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

			})

		})

		Describe("get one", func() {

			var (
				specifiedCarShare model.CarShare
				result            model.CarShare
				err               error
			)

			BeforeEach(func() {
				// select one of the existing car shares
				err = db.DB("carshare").C("carShares").Find(nil).One(&specifiedCarShare)
				Expect(err).ToNot(HaveOccurred())
				Expect(specifiedCarShare).ToNot(BeNil())
				result, err = carShareStorage.GetOne(specifiedCarShare.GetID(), context)
			})

			Context("with valid mgo connection", func() {

				Context("targeting a car share that exists", func() {

					It("should not throw an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					It("should return the specified car share", func() {
						Expect(result).To(Equal(specifiedCarShare))
					})

				})

				Context("targeting a car share that does not exists", func() {

					BeforeEach(func() {
						result, err = carShareStorage.GetOne(bson.NewObjectId().Hex(), context)
					})

					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
					})

				})

			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					result, err = carShareStorage.GetOne(specifiedCarShare.GetID(), context)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

			})

		})

		Describe("inserting", func() {

			var (
				id  string
				err error
			)

			Context("with valid mgo connection", func() {

				BeforeEach(func() {
					id, err = carShareStorage.Insert(model.CarShare{
						Name: "example car share",
					}, context)
				})

				It("should insert a new car share", func() {

					Expect(err).ToNot(HaveOccurred())
					Expect(id).ToNot(BeEmpty())

					result := model.CarShare{}
					err = db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(id)).One(&result)
					Expect(err).ToNot(HaveOccurred())
					Expect(result.GetID()).To(Equal(id))

				})
			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					id, err = carShareStorage.Insert(model.CarShare{}, context)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

			})

		})

		Describe("deleting", func() {

			var (
				err               error
				specifiedCarShare model.CarShare
			)

			BeforeEach(func() {

				// select one of the existing car shares
				err = db.DB("carshare").C("carShares").Find(nil).One(&specifiedCarShare)
				Expect(err).ToNot(HaveOccurred())
				Expect(specifiedCarShare).ToNot(BeNil())

				err = carShareStorage.Delete(specifiedCarShare.GetID(), context)
			})

			Context("with valid mgo connection", func() {

				Context("targeting a car share that exists", func() {

					It("should not throw an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					Specify("the car share should no longer exist in mongo db", func() {
						count, err := db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(specifiedCarShare.GetID())).Count()
						Expect(err).ToNot(HaveOccurred())
						Expect(count).To(BeZero())
					})

				})

				Context("targeting a car share that does not exists", func() {

					BeforeEach(func() {
						err = carShareStorage.Delete(bson.NewObjectId().Hex(), context)
					})

					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
					})

				})
			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					err = carShareStorage.Delete(bson.NewObjectId().Hex(), context)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

			})

		})

		Describe("updating", func() {

			var (
				specifiedCarShare model.CarShare
				err               error
			)

			Context("with valid mgo connection", func() {

				Context("targeting a car share that exists", func() {

					BeforeEach(func() {

						// select one of the existing car shares
						err = db.DB("carshare").C("carShares").Find(nil).One(&specifiedCarShare)
						Expect(err).ToNot(HaveOccurred())
						Expect(specifiedCarShare).ToNot(BeNil())

						// update it
						specifiedCarShare.Name = "updated"
						err = carShareStorage.Update(specifiedCarShare, context)

					})

					It("should not throw an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					Specify("the car share should be updated in mongo db", func() {
						result := model.CarShare{}
						err = db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(specifiedCarShare.GetID())).One(&result)
						Expect(err).ToNot(HaveOccurred())
						Expect(result.Name).To(Equal("updated"))
					})

				})

				Context("targeting a car share that does not exists", func() {

					BeforeEach(func() {
						err = carShareStorage.Update(model.CarShare{}, context)
					})

					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
					})

				})

			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
				})

				It("should return an error", func() {
					err := carShareStorage.Update(model.CarShare{}, context)
					Expect(err).To(HaveOccurred())
				})

			})

		})

	})
})

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
