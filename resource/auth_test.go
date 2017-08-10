package resource

import (
	"gopkg.in/jose.v1/jwt"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage/mongodb"

	"github.com/manyminds/api2go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("auth", func() {

	var (
		request = api2go.Request{
			Context: &api2go.APIContext{},
		}
		userStorage = &mongodb.UserStorage{}
		fbUser      = model.User{
			ID:          bson.NewObjectId(),
			DisplayName: "User linked to firebaseUID",
			FirebaseUID: "fbUserfirebaseuid",
		}
	)

	BeforeEach(func() {

		db, pool, containerResource = mongodb.ConnectToMongoDB(db, pool, containerResource)
		Expect(db).ToNot(BeNil())
		Expect(pool).ToNot(BeNil())
		Expect(containerResource).ToNot(BeNil())
		request.Context.Set("db", db)

		err := db.DB(mongodb.CarShareDB).DropDatabase()
		db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Insert(fbUser) // don't insert fbUser2
		Expect(err).ToNot(HaveOccurred())

		if request.Header == nil {
			request.Header = make(map[string][]string)
			request.Header.Set("authorization", "example JWT")
		}
	})

	Describe("getRequestUser", func() {

		var err error
		var requestUser model.User

		BeforeEach(func() {
			mockTokenVerifier := mockTokenVerifier{}
			mockTokenVerifier.Claims = make(jwt.Claims)
			mockTokenVerifier.Claims.Set("sub", "fbUserfirebaseuid")
			requestUser, err = getRequestUser(request, mockTokenVerifier, userStorage)
		})

		It("should throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		Context("authenticated firebase user not in users data store", func() {

			BeforeEach(func() {
				mockTokenVerifier := mockTokenVerifier{}
				mockTokenVerifier.Claims = make(jwt.Claims)
				mockTokenVerifier.Claims.Set("sub", "newUserFirebaseUID")
				requestUser, err = getRequestUser(request, mockTokenVerifier, userStorage)
			})

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return a user", func() {
				Expect(requestUser).ToNot(BeNil())
			})

			It("should return a user with the correct firebaseuid", func() {
				Expect(requestUser.FirebaseUID).To(Equal("newUserFirebaseUID"))
			})

		})

	})

})
