package mongodb_test

import (
	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	. "github.com/LewisWatson/carshare-back/storage/mongodb"

	"github.com/manyminds/api2go"
	"gopkg.in/mgo.v2/bson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Trip Storage", func() {

	var (
		tripStorage *TripStorage
		context     *api2go.APIContext
		trips       = []model.Trip{
			model.Trip{
				ID:           bson.NewObjectId(),
				Metres:       123,
				PassengerIDs: []string{},
				Scores:       map[string]model.Score{},
			},
			model.Trip{
				ID:           bson.NewObjectId(),
				Metres:       456,
				PassengerIDs: []string{},
				Scores:       map[string]model.Score{},
			},
			model.Trip{
				ID:           bson.NewObjectId(),
				Metres:       789,
				PassengerIDs: []string{},
				Scores:       map[string]model.Score{},
			},
			model.Trip{
				ID:           bson.NewObjectId(),
				Metres:       234,
				PassengerIDs: []string{},
				Scores:       map[string]model.Score{},
			},
			model.Trip{
				ID:           bson.NewObjectId(),
				Metres:       567,
				PassengerIDs: []string{},
				Scores:       map[string]model.Score{},
			},
			model.Trip{
				ID:           bson.NewObjectId(),
				Metres:       890,
				PassengerIDs: []string{},
				Scores:       map[string]model.Score{},
			},
		}
		carShares = []model.CarShare{
			model.CarShare{
				ID:   bson.NewObjectId(),
				Name: "Example Car Share 1",
				TripIDs: []string{
					trips[0].GetID(),
					trips[1].GetID(),
					trips[2].GetID(),
					trips[3].GetID(),
				},
			},
			model.CarShare{
				ID:   bson.NewObjectId(),
				Name: "Example Car Share 2",
				TripIDs: []string{
					trips[4].GetID(),
					trips[5].GetID(),
				},
			},
			model.CarShare{
				ID:   bson.NewObjectId(),
				Name: "Example Car Share 3",
			},
		}
	)

	BeforeEach(func() {
		tripStorage = &TripStorage{
			CarshareStorage: &CarShareStorage{},
		}
		context = &api2go.APIContext{}
		connectToMongoDB()
		err := db.DB("carshare").DropDatabase()
		Expect(err).ToNot(HaveOccurred())
		context.Set("db", db)
		for _, trip := range trips {
			err = db.DB("carshare").C("trips").Insert(trip)
			Expect(err).ToNot(HaveOccurred())
		}
		for _, carShare := range carShares {
			err = db.DB("carshare").C("carShares").Insert(carShare)
			Expect(err).ToNot(HaveOccurred())
		}
	})

	Describe("get all", func() {

		var (
			result []model.Trip
			err    error
		)

		BeforeEach(func() {
			result, err = tripStorage.GetAll(context)
		})

		Context("with valid mgo connection", func() {

			It("should return all existing trips", func() {
				Expect(err).ToNot(HaveOccurred())
				Ω(result).Should(ConsistOf(trips))
			})

		})

		Context("with missing mgo connection", func() {

			BeforeEach(func() {
				context.Reset()
				result, err = tripStorage.GetAll(context)
			})

			It("should return an error", func() {
				Expect(err).To(HaveOccurred())
			})

		})

	})

	Describe("get one", func() {

		var (
			result model.Trip
			err    error
		)

		Context("with valid mgo connection", func() {

			Context("targeting a trip that exists", func() {

				BeforeEach(func() {
					result, err = tripStorage.GetOne(trips[0].GetID(), context)
				})

				It("should not throw an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should return the specified trip", func() {
					trips[0].TimeStamp = trips[0].TimeStamp.UTC()
					Expect(result).To(Equal(trips[0]))
				})

			})

			Context("targeting a trip that does not exist", func() {

				Context("valid bson object id", func() {

					BeforeEach(func() {
						result, err = tripStorage.GetOne(bson.NewObjectId().Hex(), context)
					})

					It("should throw a storage.ErrNotFound error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(storage.ErrNotFound))
					})

				})

				Context("invalid bson object id", func() {

					BeforeEach(func() {
						result, err = tripStorage.GetOne("invalid id", context)
					})

					It("should throw a storage.InvalidID error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(Equal(storage.InvalidID))
					})

				})

			})

		})

		Context("with missing mgo connection", func() {

			BeforeEach(func() {
				context.Reset()
				result, err = tripStorage.GetOne(bson.NewObjectId().Hex(), context)
			})

			It("should return an ErrorNoDBSessionInContext error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrorNoDBSessionInContext))
			})

		})

	})

	// Describe("inserting", func() {

	// 	var (
	// 		specifiedCarShare model.CarShare
	// 		id                string
	// 		err               error
	// 	)

	// 	Context("with valid mgo connection", func() {

	// 		Context("targeting a car share that exists", func() {

	// 			BeforeEach(func() {

	// 				// select one of the existing car shares
	// 				err = db.DB("carshare").C("carShares").Find(nil).One(&specifiedCarShare)
	// 				Expect(err).ToNot(HaveOccurred())
	// 				Expect(specifiedCarShare).ToNot(BeNil())

	// 				id, err = tripStorage.Insert(
	// 					specifiedCarShare.GetID(),
	// 					model.Trip{
	// 						Metres: 123,
	// 					},
	// 					context,
	// 				)
	// 			})

	// 			It("should not throw an error", func() {
	// 				Expect(err).ToNot(HaveOccurred())
	// 			})

	// 			It("should result in the car share being updated with the car share", func() {

	// 				result := model.CarShare{}
	// 				err = db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(specifiedCarShare.GetID())).One(&result)
	// 				Expect(err).ToNot(HaveOccurred())

	// 				trip := model.Trip{
	// 					ID:           bson.ObjectIdHex(id),
	// 					Metres:       123,
	// 					PassengerIDs: []string{},
	// 					Scores:       map[string]model.Score{},
	// 				}

	// 				Expect(result.Trips).To(ContainElement(trip))

	// 			})

	// 		})

	// 		Context("targeting a car share that does not exist", func() {

	// 			Context("valid bson object id", func() {

	// 				BeforeEach(func() {
	// 					id, err = tripStorage.Insert(bson.NewObjectId().Hex(), model.Trip{}, context)
	// 				})

	// 				It("should throw an ErrNotFound error", func() {
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err).To(Equal(storage.ErrNotFound))
	// 				})

	// 			})

	// 			Context("invalid bson object id", func() {

	// 				BeforeEach(func() {
	// 					id, err = tripStorage.Insert("invalid id", model.Trip{}, context)
	// 				})

	// 				It("should throw an storage.InvalidID error", func() {
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err).To(Equal(storage.InvalidID))
	// 				})

	// 			})

	// 		})
	// 	})

	// 	Context("with missing mgo connection", func() {

	// 		BeforeEach(func() {
	// 			context.Reset()
	// 			id, err = tripStorage.Insert(bson.NewObjectId().Hex(), model.Trip{}, context)
	// 		})

	// 		It("should return an ErrorNoDBSessionInContext error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(ErrorNoDBSessionInContext))
	// 		})

	// 	})

	// })

	// Describe("deleting", func() {

	// 	var (
	// 		specifiedCarShare model.CarShare
	// 		id                string
	// 		err               error
	// 	)

	// 	Context("with valid mgo connection", func() {

	// 		Context("targeting a car share that exists", func() {

	// 			BeforeEach(func() {

	// 				// select one of the existing car shares
	// 				err = db.DB("carshare").C("carShares").Find(bson.M{"name": "Example Car Share 1"}).One(&specifiedCarShare)
	// 				Expect(err).ToNot(HaveOccurred())
	// 				Expect(specifiedCarShare).ToNot(BeNil())
	// 			})

	// 			Context("targeting a trip that exists in the specified car share", func() {

	// 				BeforeEach(func() {
	// 					// select one of the trips
	// 					id = "507f191e810c19729de860ea"
	// 					err = tripStorage.Delete(specifiedCarShare.GetID(), id, context)
	// 				})

	// 				It("should not throw an error", func() {
	// 					Expect(err).ToNot(HaveOccurred())
	// 				})

	// 				It("should result in the car share no longer containing the trip", func() {

	// 					result := model.CarShare{}
	// 					err = db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(specifiedCarShare.GetID())).One(&result)
	// 					Expect(err).ToNot(HaveOccurred())

	// 					_, ok := result.Trips[id]
	// 					Expect(ok).To(Equal(false))

	// 				})

	// 			})

	// 			Context("targeting a trip that does not exists in the specified car share", func() {

	// 				Context("valid bson object id", func() {

	// 					BeforeEach(func() {
	// 						err = tripStorage.Delete(specifiedCarShare.GetID(), bson.NewObjectId().Hex(), context)
	// 					})

	// 					It("should throw an storage.ErrNotFound error", func() {
	// 						Expect(err).To(HaveOccurred())
	// 						Expect(err).To(Equal(storage.ErrNotFound))
	// 					})

	// 				})

	// 				Context("invalid bson object id", func() {

	// 					BeforeEach(func() {
	// 						err = tripStorage.Delete(specifiedCarShare.GetID(), "invalid", context)
	// 					})

	// 					It("should throw an storage.InvalidID error", func() {
	// 						Expect(err).To(HaveOccurred())
	// 						Expect(err).To(Equal(storage.InvalidID))
	// 					})

	// 				})

	// 			})

	// 		})

	// 		Context("targeting a car share that does not exist", func() {

	// 			Context("valid bson object id", func() {

	// 				BeforeEach(func() {
	// 					err = tripStorage.Delete(bson.NewObjectId().Hex(), bson.NewObjectId().Hex(), context)
	// 				})

	// 				It("should throw an ErrNotFound error", func() {
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err).To(Equal(storage.ErrNotFound))
	// 				})

	// 			})

	// 			Context("invalid bson object id", func() {

	// 				BeforeEach(func() {
	// 					err = tripStorage.Delete("invalid id", bson.NewObjectId().Hex(), context)
	// 				})

	// 				It("should throw an storage.InvalidID error", func() {
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err).To(Equal(storage.InvalidID))
	// 				})

	// 			})

	// 		})
	// 	})

	// 	Context("with missing mgo connection", func() {

	// 		BeforeEach(func() {
	// 			context.Reset()
	// 			err = tripStorage.Delete(bson.NewObjectId().Hex(), bson.NewObjectId().Hex(), context)
	// 		})

	// 		It("should return an ErrorNoDBSessionInContext error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(ErrorNoDBSessionInContext))
	// 		})

	// 	})

	// })

	// Describe("updating", func() {

	// 	var (
	// 		specifiedCarShare model.CarShare
	// 		id                string
	// 		err               error
	// 	)

	// 	Context("with valid mgo connection", func() {

	// 		Context("targeting a car share that exists", func() {

	// 			BeforeEach(func() {

	// 				// select one of the existing car shares
	// 				err = db.DB("carshare").C("carShares").Find(bson.M{"name": "Example Car Share 1"}).One(&specifiedCarShare)
	// 				Expect(err).ToNot(HaveOccurred())
	// 				Expect(specifiedCarShare).ToNot(BeNil())
	// 			})

	// 			Context("targeting a trip that exists in the specified car share", func() {

	// 				BeforeEach(func() {
	// 					// select one of the trips
	// 					id = "507f191e810c19729de860ea"

	// 					trip, ok := specifiedCarShare.Trips[id]
	// 					Expect(ok).To(Equal(true))

	// 					trip.Metres = 1337
	// 					err = tripStorage.Update(specifiedCarShare.GetID(), trip, context)
	// 				})

	// 				It("should not throw an error", func() {
	// 					Expect(err).ToNot(HaveOccurred())
	// 				})

	// 				It("should result in the trip reflecting the changes", func() {

	// 					result := model.CarShare{}
	// 					err = db.DB("carshare").C("carShares").FindId(bson.ObjectIdHex(specifiedCarShare.GetID())).One(&result)
	// 					Expect(err).ToNot(HaveOccurred())

	// 					trip, ok := result.Trips[id]
	// 					Expect(ok).To(Equal(true))
	// 					Expect(trip.Metres).To(Equal(1337))

	// 				})

	// 			})

	// 			Context("targeting a trip that does not exists in the specified car share", func() {

	// 				Context("valid bson object id", func() {

	// 					BeforeEach(func() {
	// 						err = tripStorage.Update(
	// 							specifiedCarShare.GetID(),
	// 							model.Trip{
	// 								ID: bson.NewObjectId(),
	// 							},
	// 							context,
	// 						)
	// 					})

	// 					It("should throw an storage.ErrNotFound error", func() {
	// 						Expect(err).To(HaveOccurred())
	// 						Expect(err).To(Equal(storage.ErrNotFound))
	// 					})

	// 				})

	// 			})

	// 		})

	// 		Context("targeting a car share that does not exist", func() {

	// 			Context("valid bson object id", func() {

	// 				BeforeEach(func() {
	// 					err = tripStorage.Update(bson.NewObjectId().Hex(), model.Trip{}, context)
	// 				})

	// 				It("should throw an ErrNotFound error", func() {
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err).To(Equal(storage.ErrNotFound))
	// 				})

	// 			})

	// 			Context("invalid bson object id", func() {

	// 				BeforeEach(func() {
	// 					err = tripStorage.Update("invalid id", model.Trip{}, context)
	// 				})

	// 				It("should throw an storage.InvalidID error", func() {
	// 					Expect(err).To(HaveOccurred())
	// 					Expect(err).To(Equal(storage.InvalidID))
	// 				})

	// 			})

	// 		})
	// 	})

	// 	Context("with missing mgo connection", func() {

	// 		BeforeEach(func() {
	// 			context.Reset()
	// 			id, err = tripStorage.Insert(bson.NewObjectId().Hex(), model.Trip{}, context)
	// 		})

	// 		It("should return an ErrorNoDBSessionInContext error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(ErrorNoDBSessionInContext))
	// 		})

	// 	})

	// })

})
