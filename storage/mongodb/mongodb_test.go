package mongodb_test

import (
	"fmt"
	"log"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
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

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
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

				Context("targeting a car share that does not exist", func() {

					BeforeEach(func() {
						result, err = carShareStorage.GetOne(bson.NewObjectId().Hex(), context)
					})

					It("should throw an ErrNotFound error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(storage.ErrNotFound))
					})

				})

				Context("using invalid id", func() {

					BeforeEach(func() {
						result, err = carShareStorage.GetOne("invalid id", context)
					})

					It("should throw an ErrNotFound error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(storage.InvalidID))
					})

				})

			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					result, err = carShareStorage.GetOne(specifiedCarShare.GetID(), context)
				})

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
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

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
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

				Context("targeting a car share that does not exist", func() {

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

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
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

				Context("targeting a car share that does not exist", func() {

					BeforeEach(func() {
						err = carShareStorage.Update(model.CarShare{
							ID: bson.NewObjectId(),
						}, context)
					})

					It("should throw an storage.ErrNotFound error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(storage.ErrNotFound))
					})

				})

				Context("targeting a car share with invalid id", func() {

					BeforeEach(func() {
						err = carShareStorage.Update(model.CarShare{
							ID: "invalid id",
						}, context)
					})

					It("should throw an storage.InvalidID error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(storage.InvalidID))
					})

				})

			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					err = carShareStorage.Update(model.CarShare{
						ID: bson.NewObjectId(),
					}, context)
				})

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
				})

			})

		})

	})

	Describe("Trip Storage", func() {

		var (
			tripStorage *TripStorage
			context     *api2go.APIContext
		)

		BeforeEach(func() {
			tripStorage = &TripStorage{}
			context = &api2go.APIContext{}
			connectToMongoDB()
			err := db.DB("carshare").DropDatabase()
			Expect(err).ToNot(HaveOccurred())
			context.Set("db", db)

			err = db.DB("carshare").C("carShares").Insert(
				model.CarShare{
					Name: "Example Car Share 1",
					Trips: map[string]model.Trip{
						"507f191e810c19729de860ea": model.Trip{
							ID: bson.ObjectIdHex("507f191e810c19729de860ea"),
						},
						"507f191e810c19729de860eb": model.Trip{
							ID: bson.ObjectIdHex("507f191e810c19729de860eb"),
						},
						"507f191e810c19729de860ec": model.Trip{
							ID: bson.ObjectIdHex("507f191e810c19729de860ec"),
						},
					},
				},
				model.CarShare{
					Name: "Example Car Share 2",
					Trips: map[string]model.Trip{
						"507f191e810c19729de860ed": model.Trip{
							ID: bson.ObjectIdHex("507f191e810c19729de860ed"),
						},
						"507f191e810c19729de860ee": model.Trip{
							ID: bson.ObjectIdHex("507f191e810c19729de860ee"),
						},
					},
				},
				model.CarShare{
					Name: "Example Car Share 3",
				},
			)
			Expect(err).ToNot(HaveOccurred())
		})

		Describe("get all", func() {

			var (
				specifiedCarShare model.CarShare
				result            map[string]model.Trip
				err               error
			)

			BeforeEach(func() {
				// select one of the existing car shares
				err = db.DB("carshare").C("carShares").Find(nil).One(&specifiedCarShare)
				Expect(err).ToNot(HaveOccurred())
				Expect(specifiedCarShare).ToNot(BeNil())

				result, err = tripStorage.GetAll(specifiedCarShare.GetID(), context)
			})

			Context("with valid mgo connection", func() {

				Context("targeting existing car share", func() {

					It("should return all existing trips for the specified car share", func() {
						Expect(err).ToNot(HaveOccurred())
						Expect(result).To(HaveKeyWithValue("507f191e810c19729de860ea", specifiedCarShare.Trips["507f191e810c19729de860ea"]))
						Expect(result).To(HaveKeyWithValue("507f191e810c19729de860eb", specifiedCarShare.Trips["507f191e810c19729de860eb"]))
						Expect(result).To(HaveKeyWithValue("507f191e810c19729de860ec", specifiedCarShare.Trips["507f191e810c19729de860ec"]))
					})

				})

				Context("targeting non existing car share", func() {

					BeforeEach(func() {
						result, err = tripStorage.GetAll(bson.NewObjectId().Hex(), context)
					})

					It("should return storage.ErrNotFound error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(storage.ErrNotFound))
					})

				})

				Context("with invalid car share id", func() {

					BeforeEach(func() {
						result, err = tripStorage.GetAll("invalid id", context)
					})

					It("should return an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(storage.InvalidID))
					})

				})

			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					result, err = tripStorage.GetAll(specifiedCarShare.GetID(), context)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

			})

		})

		Describe("get one", func() {

			var (
				specifiedCarShare model.CarShare
				specifiedTrip     model.Trip
				result            model.Trip
				err               error
			)

			Context("with valid mgo connection", func() {

				Context("targeting a car share that exists", func() {

					BeforeEach(func() {

						// select one of the existing car shares
						err = db.DB("carshare").C("carShares").Find(bson.M{"name": "Example Car Share 1"}).One(&specifiedCarShare)
						Expect(err).ToNot(HaveOccurred())
						Expect(specifiedCarShare).ToNot(BeNil())

						// select one of the trips
						var ok bool
						specifiedTrip, ok = specifiedCarShare.Trips["507f191e810c19729de860ea"]
						Expect(ok).To(Equal(true))

						// simulate the backend setting the timestamp to UTC
						specifiedTrip.TimeStamp = specifiedTrip.TimeStamp.UTC()

					})

					Context("targeting a trip that exists in the car share", func() {

						BeforeEach(func() {
							result, err = tripStorage.GetOne(specifiedCarShare.GetID(), specifiedTrip.GetID(), context)
						})

						It("should not throw an error", func() {
							Expect(err).ToNot(HaveOccurred())
						})

						It("should return the specified trip", func() {
							Expect(result).To(Equal(specifiedTrip))
						})

					})

					Context("targeting a trip that does not exists in the car share", func() {

						Context("valid bson object id", func() {

							BeforeEach(func() {
								result, err = tripStorage.GetOne(specifiedCarShare.GetID(), bson.NewObjectId().Hex(), context)
							})

							It("should throw a storage.ErrNotFound error", func() {
								Expect(err).To(HaveOccurred())
								Expect(err).To(Equal(storage.ErrNotFound))
							})

						})

						Context("invalid bson object id", func() {

							BeforeEach(func() {
								result, err = tripStorage.GetOne(specifiedCarShare.GetID(), "invalid id", context)
							})

							It("should throw a storage.InvalidID error", func() {
								Expect(err).To(HaveOccurred())
								Expect(err).To(Equal(storage.InvalidID))
							})

						})

					})

				})

				Context("targeting a car share that does not exist", func() {

					Context("valid bson object id", func() {

						BeforeEach(func() {
							result, err = tripStorage.GetOne(bson.NewObjectId().Hex(), bson.NewObjectId().Hex(), context)
						})

						It("should throw an ErrNotFound error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(Equal(storage.ErrNotFound))
						})

					})

					Context("invalid bson object id", func() {

						BeforeEach(func() {
							result, err = tripStorage.GetOne("invalid id", bson.NewObjectId().Hex(), context)
						})

						It("should throw an storage.InvalidID error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(Equal(storage.InvalidID))
						})

					})

				})

			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					result, err = tripStorage.GetOne(bson.NewObjectId().Hex(), bson.NewObjectId().Hex(), context)
				})

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
				})

			})

		})

		Describe("inserting", func() {

			var (
				specifiedCarShare model.CarShare
				id                string
				err               error
			)

			Context("with valid mgo connection", func() {

				Context("targeting a car share that exists", func() {

					BeforeEach(func() {

						// select one of the existing car shares
						err = db.DB("carshare").C("carShares").Find(nil).One(&specifiedCarShare)
						Expect(err).ToNot(HaveOccurred())
						Expect(specifiedCarShare).ToNot(BeNil())

						id, err = tripStorage.Insert(
							specifiedCarShare.GetID(),
							model.Trip{
								Metres: 123,
							},
							context,
						)
					})

					It("should not throw an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					It("should result in the car share being updated with the car share", func() {

						result := model.CarShare{}
						err = db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(specifiedCarShare.GetID())).One(&result)
						Expect(err).ToNot(HaveOccurred())

						trip := model.Trip{
							ID:           bson.ObjectIdHex(id),
							Metres:       123,
							PassengerIDs: []string{},
							Scores:       map[string]model.Score{},
						}

						Expect(result.Trips).To(ContainElement(trip))

					})

				})

				Context("targeting a car share that does not exist", func() {

					Context("valid bson object id", func() {

						BeforeEach(func() {
							id, err = tripStorage.Insert(bson.NewObjectId().Hex(), model.Trip{}, context)
						})

						It("should throw an ErrNotFound error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(Equal(storage.ErrNotFound))
						})

					})

					Context("invalid bson object id", func() {

						BeforeEach(func() {
							id, err = tripStorage.Insert("invalid id", model.Trip{}, context)
						})

						It("should throw an storage.InvalidID error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(Equal(storage.InvalidID))
						})

					})

				})
			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					id, err = tripStorage.Insert(bson.NewObjectId().Hex(), model.Trip{}, context)
				})

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
				})

			})

		})

		Describe("deleting", func() {

			var (
				specifiedCarShare model.CarShare
				id                string
				err               error
			)

			Context("with valid mgo connection", func() {

				Context("targeting a car share that exists", func() {

					BeforeEach(func() {

						// select one of the existing car shares
						err = db.DB("carshare").C("carShares").Find(bson.M{"name": "Example Car Share 1"}).One(&specifiedCarShare)
						Expect(err).ToNot(HaveOccurred())
						Expect(specifiedCarShare).ToNot(BeNil())
					})

					Context("targeting a trip that exists in the specified car share", func() {

						BeforeEach(func() {
							// select one of the trips
							id = "507f191e810c19729de860ea"
							err = tripStorage.Delete(specifiedCarShare.GetID(), id, context)
						})

						It("should not throw an error", func() {
							Expect(err).ToNot(HaveOccurred())
						})

						It("should result in the car share no longer containing the trip", func() {

							result := model.CarShare{}
							err = db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(specifiedCarShare.GetID())).One(&result)
							Expect(err).ToNot(HaveOccurred())

							_, ok := result.Trips[id]
							Expect(ok).To(Equal(false))

						})

					})

					Context("targeting a trip that does not exists in the specified car share", func() {

						Context("valid bson object id", func() {

							BeforeEach(func() {
								err = tripStorage.Delete(specifiedCarShare.GetID(), bson.NewObjectId().Hex(), context)
							})

							It("should throw an storage.ErrNotFound error", func() {
								Expect(err).To(HaveOccurred())
								Expect(err).To(Equal(storage.ErrNotFound))
							})

						})

						Context("invalid bson object id", func() {

							BeforeEach(func() {
								err = tripStorage.Delete(specifiedCarShare.GetID(), "invalid", context)
							})

							It("should throw an storage.InvalidID error", func() {
								Expect(err).To(HaveOccurred())
								Expect(err).To(Equal(storage.InvalidID))
							})

						})

					})

				})

				Context("targeting a car share that does not exist", func() {

					Context("valid bson object id", func() {

						BeforeEach(func() {
							err = tripStorage.Delete(bson.NewObjectId().Hex(), bson.NewObjectId().Hex(), context)
						})

						It("should throw an ErrNotFound error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(Equal(storage.ErrNotFound))
						})

					})

					Context("invalid bson object id", func() {

						BeforeEach(func() {
							err = tripStorage.Delete("invalid id", bson.NewObjectId().Hex(), context)
						})

						It("should throw an storage.InvalidID error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(Equal(storage.InvalidID))
						})

					})

				})
			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					err = tripStorage.Delete(bson.NewObjectId().Hex(), bson.NewObjectId().Hex(), context)
				})

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
				})

			})

		})

		Describe("updating", func() {

			var (
				specifiedCarShare model.CarShare
				id                string
				err               error
			)

			Context("with valid mgo connection", func() {

				Context("targeting a car share that exists", func() {

					BeforeEach(func() {

						// select one of the existing car shares
						err = db.DB("carshare").C("carShares").Find(bson.M{"name": "Example Car Share 1"}).One(&specifiedCarShare)
						Expect(err).ToNot(HaveOccurred())
						Expect(specifiedCarShare).ToNot(BeNil())
					})

					Context("targeting a trip that exists in the specified car share", func() {

						BeforeEach(func() {
							// select one of the trips
							id = "507f191e810c19729de860ea"

							trip, ok := specifiedCarShare.Trips[id]
							Expect(ok).To(Equal(true))

							trip.Metres = 1337
							err = tripStorage.Update(specifiedCarShare.GetID(), trip, context)
						})

						It("should not throw an error", func() {
							Expect(err).ToNot(HaveOccurred())
						})

						It("should result in the trip reflecting the changes", func() {

							result := model.CarShare{}
							err = db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(specifiedCarShare.GetID())).One(&result)
							Expect(err).ToNot(HaveOccurred())

							trip, ok := result.Trips[id]
							Expect(ok).To(Equal(true))
							Expect(trip.Metres).To(Equal(1337))

						})

					})

					Context("targeting a trip that does not exists in the specified car share", func() {

						Context("valid bson object id", func() {

							BeforeEach(func() {
								err = tripStorage.Update(
									specifiedCarShare.GetID(),
									model.Trip{
										ID: bson.NewObjectId(),
									},
									context,
								)
							})

							It("should throw an storage.ErrNotFound error", func() {
								Expect(err).To(HaveOccurred())
								Expect(err).To(Equal(storage.ErrNotFound))
							})

						})

					})

				})

				Context("targeting a car share that does not exist", func() {

					Context("valid bson object id", func() {

						BeforeEach(func() {
							err = tripStorage.Update(bson.NewObjectId().Hex(), model.Trip{}, context)
						})

						It("should throw an ErrNotFound error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(Equal(storage.ErrNotFound))
						})

					})

					Context("invalid bson object id", func() {

						BeforeEach(func() {
							err = tripStorage.Update("invalid id", model.Trip{}, context)
						})

						It("should throw an storage.InvalidID error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(Equal(storage.InvalidID))
						})

					})

				})
			})

			Context("with missing mgo connection", func() {

				BeforeEach(func() {
					context.Reset()
					id, err = tripStorage.Insert(bson.NewObjectId().Hex(), model.Trip{}, context)
				})

				It("should return an ErrorNoDBSessionInContext error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(ErrorNoDBSessionInContext))
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
