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

var _ = Describe("Car Share Storage", func() {

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
