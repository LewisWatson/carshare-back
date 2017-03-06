package resource

import (
	"fmt"
	"net/http"

	"gopkg.in/jose.v1/jwt"

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
		request.Header = make(map[string][]string)
		request.Header.Set("authorization", "example JWT")
		db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Insert(
			&model.User{
				DisplayName: "Example User 1",
			},
			&model.User{
				DisplayName: "Example User 2",
			},
		)
	})

	Describe("get all", func() {

		var err error

		BeforeEach(func() {
			_, err = userResource.FindAll(request)
		})

		It("should throw an error", func() {
			Expect(err).To(HaveOccurred())
		})

	})

	Describe("get one", func() {

		var err error

		BeforeEach(func() {
			_, err = userResource.FindOne("", request)
		})

		It("should throw an error", func() {
			Expect(err).To(HaveOccurred())
		})

	})

	Describe("create", func() {

		var (
			user = model.User{
				FirebaseUID: "example firebase UID",
				DisplayName: "example",
				Email:       "user@example.com",
				PhotoURL:    "http://photo.org",
				IsAnon:      false,
			}
			result api2go.Responder
			err    error
		)

		BeforeEach(func() {

			// simulate the request coming in with a valid JWT token for the
			// user being created
			mockTokenVerifier := mockTokenVerifier{}
			mockTokenVerifier.Claims = make(jwt.Claims)
			mockTokenVerifier.Claims.Set("sub", user.FirebaseUID)
			userResource.TokenVerifier = mockTokenVerifier

			result, err = userResource.Create(user, request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return http status created", func() {
			Expect(result.StatusCode()).To(Equal(http.StatusCreated))
		})

		It("should persist and return the user", func() {
			Expect(result.Result()).To(BeAssignableToTypeOf(model.User{}))
			resUser := result.Result().(model.User)
			user.ID = resUser.ID
			Expect(resUser).To(Equal(user))
			persistedUser, err := userResource.UserStorage.GetOne(resUser.GetID(), context)
			Expect(err).ToNot(HaveOccurred())
			Expect(persistedUser).To(Equal(user))
		})

		Context("user not logged in", func() {

			BeforeEach(func() {
				mockTokenVerifier := userResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Error = fmt.Errorf("example error")
				userResource.TokenVerifier = mockTokenVerifier
				result, err = userResource.Create(user, request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, Error creating user, example error"))
			})

		})

		Context("attempt to create a user for a different firebase user", func() {

			var tokenSub = "aDifferentFirebaseUser"

			BeforeEach(func() {
				mockTokenVerifier := mockTokenVerifier{}
				mockTokenVerifier.Claims = make(jwt.Claims)
				mockTokenVerifier.Claims.Set("sub", tokenSub)
				userResource.TokenVerifier = mockTokenVerifier

				result, err = userResource.Create(user, request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) You cannot create a user for another firebase user and 0 more errors, FirebaseUID \"" + tokenSub + "\" attempting to create user with FirebaseUID \"" + user.FirebaseUID + "\""))
			})

		})

		Context("create a user not associated with firebase user", func() {

			var user2 = model.User{
				DisplayName: "User not linked to account",
			}

			BeforeEach(func() {
				result, err = userResource.Create(user2, request)
			})

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return http status created", func() {
				Expect(result.StatusCode()).To(Equal(http.StatusCreated))
			})

			It("should persist and return the user", func() {
				Expect(result.Result()).To(BeAssignableToTypeOf(model.User{}))
				resUser := result.Result().(model.User)
				user2.ID = resUser.ID
				Expect(resUser).To(Equal(user2))
				persistedUser, err := userResource.UserStorage.GetOne(resUser.GetID(), context)
				Expect(err).ToNot(HaveOccurred())
				Expect(persistedUser).To(Equal(user2))
			})

		})

	})

})
