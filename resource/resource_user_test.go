package resource

import (
	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage/mongodb"

	"github.com/manyminds/api2go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Resource", func() {

	var (
		userResource *UserResource
		request      api2go.Request
		context      *api2go.APIContext
	)

	BeforeEach(func() {
		userResource = &UserResource{
			UserStorage: &mongodb.UserStorage{},
		}
		context = &api2go.APIContext{}
		db, pool, containerResource = mongodb.ConnectToMongoDB(db, pool, containerResource)
		Expect(db).ToNot(BeNil())
		Expect(pool).ToNot(BeNil())
		Expect(containerResource).ToNot(BeNil())
		err := db.DB(mongodb.CarShareDB).DropDatabase()
		Expect(err).ToNot(HaveOccurred())
		context.Set("db", db)
		request = api2go.Request{
			Context: context,
		}
		db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Insert(
			&model.User{
				Username: "Example User 1",
			},
			&model.User{
				Username: "Example User 2",
			},
		)
	})

	Describe("get all", func() {

		var (
			users  []model.User
			result api2go.Responder
			err    error
		)

		BeforeEach(func() {
			users, err = userResource.UserStorage.GetAll(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(users).NotTo(BeNil())
			result, err = userResource.FindAll(request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return all existing users", func() {
			Expect(result).ToNot(BeNil())
			response, ok := result.(*Response)
			Expect(ok).To(Equal(true))
			Expect(response.Res).To(Equal(users))
		})

	})

	// Describe("get one", func() {

	// 	var (
	// 		specifiedUser model.User
	// 		result        model.User
	// 		err           error
	// 	)

	// 	BeforeEach(func() {
	// 		// select one of the existing users
	// 		err = db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Find(nil).One(&specifiedUser)
	// 		Expect(err).ToNot(HaveOccurred())
	// 		Expect(specifiedUser).ToNot(BeNil())
	// 		result, err = userStorage.GetOne(specifiedUser.GetID(), context)
	// 	})

	// 	Context("targeting a user that exists", func() {

	// 		It("should not throw an error", func() {
	// 			Expect(err).ToNot(HaveOccurred())
	// 		})

	// 		It("should return the specified user", func() {
	// 			Expect(result).To(Equal(specifiedUser))
	// 		})

	// 	})

	// 	Context("targeting a user that does not exist", func() {

	// 		BeforeEach(func() {
	// 			result, err = userStorage.GetOne(bson.NewObjectId().Hex(), context)
	// 		})

	// 		It("should throw an ErrNotFound error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(storage.ErrNotFound))
	// 		})

	// 	})

	// 	Context("using invalid id", func() {

	// 		BeforeEach(func() {
	// 			result, err = userStorage.GetOne("invalid id", context)
	// 		})

	// 		It("should throw an ErrNotFound error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(storage.InvalidID))
	// 		})

	// 	})

	// 	Context("with missing mgo connection", func() {

	// 		BeforeEach(func() {
	// 			context.Reset()
	// 			result, err = userStorage.GetOne(specifiedUser.GetID(), context)
	// 		})

	// 		It("should return an ErrorNoDBSessionInContext error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(ErrorNoDBSessionInContext))
	// 		})

	// 	})

	// })

	// Describe("inserting", func() {

	// 	var (
	// 		id  string
	// 		err error
	// 	)

	// 	BeforeEach(func() {
	// 		id, err = userStorage.Insert(model.User{
	// 			Username: "example user",
	// 		}, context)
	// 	})

	// 	It("should insert a new user", func() {

	// 		Expect(err).ToNot(HaveOccurred())
	// 		Expect(id).ToNot(BeEmpty())

	// 		result := model.User{}
	// 		err = db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).FindId(bson.ObjectIdHex(id)).One(&result)
	// 		Expect(err).ToNot(HaveOccurred())
	// 		Expect(result.GetID()).To(Equal(id))

	// 	})

	// 	Context("with missing mgo connection", func() {

	// 		BeforeEach(func() {
	// 			context.Reset()
	// 			id, err = userStorage.Insert(model.User{}, context)
	// 		})

	// 		It("should return an ErrorNoDBSessionInContext error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(ErrorNoDBSessionInContext))
	// 		})

	// 	})

	// })

	// Describe("deleting", func() {

	// 	var (
	// 		err           error
	// 		specifiedUser model.User
	// 	)

	// 	BeforeEach(func() {

	// 		// select one of the existing users
	// 		err = db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Find(nil).One(&specifiedUser)
	// 		Expect(err).ToNot(HaveOccurred())
	// 		Expect(specifiedUser).ToNot(BeNil())

	// 		err = userStorage.Delete(specifiedUser.GetID(), context)
	// 	})

	// 	Context("targeting a user that exists", func() {

	// 		It("should not throw an error", func() {
	// 			Expect(err).ToNot(HaveOccurred())
	// 		})

	// 		Specify("the user should no longer exist in mongo db", func() {
	// 			count, err := db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).FindId(bson.ObjectIdHex(specifiedUser.GetID())).Count()
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(count).To(BeZero())
	// 		})

	// 	})

	// 	Context("targeting a user that does not exist", func() {

	// 		BeforeEach(func() {
	// 			err = userStorage.Delete(bson.NewObjectId().Hex(), context)
	// 		})

	// 		It("should throw an error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 		})

	// 	})

	// 	Context("with missing mgo connection", func() {

	// 		BeforeEach(func() {
	// 			context.Reset()
	// 			err = userStorage.Delete(bson.NewObjectId().Hex(), context)
	// 		})

	// 		It("should return an ErrorNoDBSessionInContext error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(ErrorNoDBSessionInContext))
	// 		})

	// 	})

	// })

	// Describe("updating", func() {

	// 	var (
	// 		specifiedUser model.User
	// 		err           error
	// 	)

	// 	Context("targeting a user that exists", func() {

	// 		BeforeEach(func() {

	// 			// select one of the existing users
	// 			err = db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Find(nil).One(&specifiedUser)
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(specifiedUser).ToNot(BeNil())

	// 			// update it
	// 			specifiedUser.Username = "updated"
	// 			err = userStorage.Update(specifiedUser, context)

	// 		})

	// 		It("should not throw an error", func() {
	// 			Expect(err).ToNot(HaveOccurred())
	// 		})

	// 		Specify("the user should be updated in mongo db", func() {
	// 			result := model.User{}
	// 			err = db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).FindId(bson.ObjectIdHex(specifiedUser.GetID())).One(&result)
	// 			Expect(err).ToNot(HaveOccurred())
	// 			Expect(result.Username).To(Equal("updated"))
	// 		})

	// 	})

	// 	Context("targeting a user that does not exist", func() {

	// 		BeforeEach(func() {
	// 			err = userStorage.Update(model.User{
	// 				ID: bson.NewObjectId(),
	// 			}, context)
	// 		})

	// 		It("should throw an storage.ErrNotFound error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(storage.ErrNotFound))
	// 		})

	// 	})

	// 	Context("targeting a user with invalid id", func() {

	// 		BeforeEach(func() {
	// 			err = userStorage.Update(model.User{
	// 				ID: "invalid id",
	// 			}, context)
	// 		})

	// 		It("should throw an storage.InvalidID error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(storage.InvalidID))
	// 		})

	// 	})

	// 	Context("with missing mgo connection", func() {

	// 		BeforeEach(func() {
	// 			context.Reset()
	// 			err = userStorage.Update(model.User{
	// 				ID: bson.NewObjectId(),
	// 			}, context)
	// 		})

	// 		It("should return an ErrorNoDBSessionInContext error", func() {
	// 			Expect(err).To(HaveOccurred())
	// 			Expect(err).To(Equal(ErrorNoDBSessionInContext))
	// 		})

	// 	})

	// })

})