package mongodb

import (
	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"

	"github.com/manyminds/api2go"
	mgo "gopkg.in/mgo.v2"
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
		}
	)

	BeforeEach(func() {
		tripStorage = &TripStorage{
			CarshareStorage: &CarShareStorage{},
		}
		context = &api2go.APIContext{}
		db, pool, containerResource = ConnectToMongoDB(db, pool, containerResource)
		err := db.DB(CarShareDB).DropDatabase()
		Expect(err).ToNot(HaveOccurred())
		context.Set("db", db)
		for _, trip := range trips {
			err = db.DB(CarShareDB).C(TripsColl).Insert(trip)
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
				Ω(result).Should(ContainElement(trips[0]))
				Ω(result).Should(ContainElement(trips[1]))
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

	Describe("inserting", func() {

		var (
			id  string
			err error
		)

		BeforeEach(func() {
			id, err = tripStorage.Insert(
				model.Trip{
					Metres: 123,
				},
				context,
			)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should result in the trip appearing in the database", func() {

			result := model.Trip{}
			err = db.DB(CarShareDB).C(TripsColl).FindId(bson.ObjectIdHex(id)).One(&result)
			Expect(err).ToNot(HaveOccurred())
			Expect(result.Metres).To(Equal(123))

		})

		Context("with missing mgo connection", func() {

			BeforeEach(func() {
				context.Reset()
				id, err = tripStorage.Insert(model.Trip{}, context)
			})

			It("should return an ErrorNoDBSessionInContext error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrorNoDBSessionInContext))
			})

		})

	})

	Describe("deleting", func() {

		var err error

		Context("targeting a trip that exists", func() {

			BeforeEach(func() {
				err = tripStorage.Delete(trips[0].GetID(), context)
			})

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should delete the trip", func() {
				result := model.Trip{}
				err = db.DB(CarShareDB).C(TripsColl).FindId(bson.ObjectIdHex(trips[0].GetID())).One(&result)
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(mgo.ErrNotFound))
			})

		})

		Context("targeting a trip that does not exists", func() {

			Context("valid bson object id", func() {

				BeforeEach(func() {
					err = tripStorage.Delete(bson.NewObjectId().Hex(), context)
				})

				It("should throw an storage.ErrNotFound error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err).To(Equal(storage.ErrNotFound))
				})

			})

			Context("invalid bson object id", func() {

				BeforeEach(func() {
					err = tripStorage.Delete("invalid", context)
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
				err = tripStorage.Delete(bson.NewObjectId().Hex(), context)
			})

			It("should return an ErrorNoDBSessionInContext error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrorNoDBSessionInContext))
			})

		})

	})

	Describe("updating", func() {

		var (
			id  string
			err error
		)

		Context("targeting a trip that exists", func() {

			BeforeEach(func() {
				trips[0].Metres = 1337
				err = tripStorage.Update(trips[0], context)
			})

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should result in the trip reflecting the changes", func() {
				result := model.Trip{}
				err = db.DB(CarShareDB).C(TripsColl).FindId(bson.ObjectIdHex(trips[0].GetID())).One(&result)
				Expect(err).ToNot(HaveOccurred())
				Expect(result.Metres).To(Equal(1337))
			})

		})

		Context("targeting a trip that does not exist", func() {

			Context("valid bson object id", func() {

				BeforeEach(func() {
					err = tripStorage.Update(
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

		Context("with missing mgo connection", func() {

			BeforeEach(func() {
				context.Reset()
				id, err = tripStorage.Insert(model.Trip{}, context)
			})

			It("should return an ErrorNoDBSessionInContext error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err).To(Equal(ErrorNoDBSessionInContext))
			})

		})

	})

})
