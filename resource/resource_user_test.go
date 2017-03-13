package resource

import (
	"fmt"
	"log"
	"net/http"

	"gopkg.in/jose.v1/jwt"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage/mongodb"

	"github.com/manyminds/api2go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("User Resource", func() {

	var (
		request = api2go.Request{
			Context: &api2go.APIContext{},
		}
		userResource = &UserResource{
			UserStorage:     &mongodb.UserStorage{},
			CarShareStorage: &mongodb.CarShareStorage{},
		}
		carShare = model.CarShare{
			ID:   bson.NewObjectId(),
			Name: "Example car share",
		}
		fbUser = model.User{
			ID:          bson.NewObjectId(),
			DisplayName: "User linked to firebaseUID",
			FirebaseUID: "fbUserfirebaseuid",
		}
		csLinkedUser = model.User{
			ID:               bson.NewObjectId(),
			DisplayName:      "User linked to car share " + carShare.GetID(),
			LinkedCarShareID: carShare.GetID(),
		}
	)

	BeforeEach(func() {

		db, pool, containerResource = mongodb.ConnectToMongoDB(db, pool, containerResource)
		Expect(db).ToNot(BeNil())
		Expect(pool).ToNot(BeNil())
		Expect(containerResource).ToNot(BeNil())
		request.Context.Set("db", db)

		err := db.DB(mongodb.CarShareDB).DropDatabase()
		db.DB(mongodb.CarShareDB).C(mongodb.CarSharesColl).Insert(carShare)
		Expect(err).ToNot(HaveOccurred())
		db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Insert(fbUser, csLinkedUser)
		Expect(err).ToNot(HaveOccurred())

		if request.Header == nil {
			request.Header = make(map[string][]string)
			request.Header.Set("authorization", "example JWT")
		}
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
			persistedUser, err := userResource.UserStorage.GetOne(resUser.GetID(), request.Context)
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

		Context("user with token for firebase user x attempts to create a user for firebase user y", func() {

			var tokenSub = "aDifferentFirebaseUser"

			BeforeEach(func() {
				mockTokenVerifier := mockTokenVerifier{}
				mockTokenVerifier.Claims = make(jwt.Claims)
				mockTokenVerifier.Claims.Set("sub", tokenSub)
				userResource.TokenVerifier = mockTokenVerifier

				result, err = userResource.Create(user, request)
			})

			It("should return a 400 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (400) Cannot create a user linked to another firebase user and 0 more errors, FirebaseUID \"" + tokenSub + "\" attempting to create user with FirebaseUID \"" + user.FirebaseUID + "\""))
			})

		})

		Context("create a user specifically for a car share", func() {

			var csLinkedUser2 = model.User{
				DisplayName:      "another user linked to car share " + carShare.GetID(),
				LinkedCarShareID: carShare.GetID(),
			}

			Context("existing car share", func() {

				BeforeEach(func() {
					result, err = userResource.Create(csLinkedUser2, request)
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
					csLinkedUser2.ID = resUser.ID
					Expect(resUser).To(Equal(csLinkedUser2))
					persistedUser, err := userResource.UserStorage.GetOne(resUser.GetID(), request.Context)
					Expect(err).ToNot(HaveOccurred())
					Expect(persistedUser).To(Equal(csLinkedUser2))
				})

			})

			Context("non-existing car share", func() {

				BeforeEach(func() {
					result, err = userResource.Create(csLinkedUser, request)
				})

				It("should throw an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})
			})
		})

		Context("not associated with either firebase or car share", func() {

			BeforeEach(func() {
				result, err = userResource.Create(model.User{}, request)
			})

			It("should throw an error", func() {
				Expect(err).To(HaveOccurred())
			})
		})

	})

	Describe("update", func() {

		var (
			result api2go.Responder
			err    error
		)

		BeforeEach(func() {

			// simulate the request coming in with a valid JWT token for the
			// user being created
			mockTokenVerifier := mockTokenVerifier{}
			mockTokenVerifier.Claims = make(jwt.Claims)
			mockTokenVerifier.Claims.Set("sub", fbUser.FirebaseUID)
			userResource.TokenVerifier = mockTokenVerifier

			fbUser.DisplayName = "updated"

			result, err = userResource.Update(fbUser, request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return http status created", func() {
			Expect(result.StatusCode()).To(Equal(http.StatusNoContent))
		})

		It("should persist and return the user", func() {
			Expect(result.Result()).To(BeAssignableToTypeOf(model.User{}))
			resUser := result.Result().(model.User)
			fbUser.ID = resUser.ID
			Expect(resUser).To(Equal(fbUser))
			persistedUser, err := userResource.UserStorage.GetOne(resUser.GetID(), request.Context)
			Expect(err).ToNot(HaveOccurred())
			Expect(persistedUser).To(Equal(fbUser))
		})

		Context("user not logged in", func() {

			BeforeEach(func() {
				mockTokenVerifier := userResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Error = fmt.Errorf("example error")
				userResource.TokenVerifier = mockTokenVerifier
				result, err = userResource.Update(fbUser, request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, Error updating user, example error"))
			})

		})

		Context("attempt to update a user for a different firebase user", func() {

			var tokenSub = "aDifferentFirebaseUser"

			BeforeEach(func() {
				mockTokenVerifier := mockTokenVerifier{}
				mockTokenVerifier.Claims = make(jwt.Claims)
				mockTokenVerifier.Claims.Set("sub", tokenSub)
				userResource.TokenVerifier = mockTokenVerifier

				result, err = userResource.Update(fbUser, request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) cannot update a user for another firebase user and 0 more errors, FirebaseUID \"" + tokenSub + "\" attempting to update user with FirebaseUID \"" + fbUser.FirebaseUID + "\""))
			})

		})

		Context("update a user not associated with firebase", func() {

			BeforeEach(func() {
				csLinkedUser.DisplayName = csLinkedUser.DisplayName + " updated"
				result, err = userResource.Update(csLinkedUser, request)
			})

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return a result", func() {
				Expect(result).ToNot(BeNil())
			})

			It("should return http status no content", func() {
				Expect(result.StatusCode()).To(Equal(http.StatusNoContent))
			})

			It("should persist and return the user", func() {
				log.Printf("checking result")
				Expect(result.Result()).To(BeAssignableToTypeOf(model.User{}))
				log.Printf("done")
				resUser := result.Result().(model.User)
				csLinkedUser.ID = resUser.ID
				log.Printf("comparing resUser to csLinkedUser")
				Expect(resUser).To(Equal(csLinkedUser))
				persistedUser, err := userResource.UserStorage.GetOne(resUser.GetID(), request.Context)
				Expect(err).ToNot(HaveOccurred())
				Expect(persistedUser).To(Equal(csLinkedUser))
			})

		})

	})

})
