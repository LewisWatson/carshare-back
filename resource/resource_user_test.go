package resource

import (
	"fmt"
	"net/http"

	"gopkg.in/jose.v1/jwt"
	"gopkg.in/mgo.v2/bson"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
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
		fbUser = model.User{
			ID:          bson.NewObjectId(),
			DisplayName: "User linked to firebaseUID",
			FirebaseUID: "fbUserfirebaseuid",
		}
		csLinkedUserID = bson.NewObjectId()
		carShare       = model.CarShare{
			ID:        bson.NewObjectId(),
			Name:      "Example car share",
			AdminIDs:  []string{fbUser.GetID()},
			MemberIDs: []string{csLinkedUserID.Hex()},
		}
		csLinkedUser = model.User{
			ID:               csLinkedUserID,
			DisplayName:      "User linked to car share " + carShare.GetID(),
			LinkedCarShareID: carShare.GetID(),
		}
		fbUser2 = model.User{
			ID:          bson.NewObjectId(),
			DisplayName: "User2 linked to firebaseUID",
			FirebaseUID: "fbUser2firebaseuid",
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
		db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Insert(fbUser, csLinkedUser, fbUser2)
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

		Context("with just a firebase UID", func() {

			BeforeEach(func() {

				user = model.User{
					FirebaseUID: "example firebase UID",
				}

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
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, error creating user: example error"))
			})

		})

		Context("user with token for firebase user x attempts to create a user for firebase user y", func() {

			BeforeEach(func() {

				// simulate being authenticated as fbUser which is admin for carshare
				mockTokenVerifier := mockTokenVerifier{}
				mockTokenVerifier.Claims = make(jwt.Claims)
				mockTokenVerifier.Claims.Set("sub", fbUser.FirebaseUID)
				userResource.TokenVerifier = mockTokenVerifier

				result, err = userResource.Create(user, request)
			})

			It("should return a 400 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) cannot create/update a user associated with another firebase user and 0 more errors, error creating user, user " + fbUser.GetID() + " (firebaseUID " + fbUser.FirebaseUID + ") attempting to create/update user " + user.GetID() + " (firebaseUID " + user.FirebaseUID + ")"))
			})

		})

		Context("create a user specifically for a car share", func() {

			var csLinkedUser2 = model.User{
				DisplayName:      "another user linked to car share " + carShare.GetID(),
				LinkedCarShareID: carShare.GetID(),
			}

			Context("existing car share", func() {

				BeforeEach(func() {

					// simulate the request coming in with a valid JWT token for the
					// user being created
					mockTokenVerifier := mockTokenVerifier{}
					mockTokenVerifier.Claims = make(jwt.Claims)
					mockTokenVerifier.Claims.Set("sub", fbUser.FirebaseUID)
					userResource.TokenVerifier = mockTokenVerifier

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
					Expect(err).To(HaveOccurred())
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

			BeforeEach(func() {
				mockTokenVerifier := mockTokenVerifier{}
				mockTokenVerifier.Claims = make(jwt.Claims)
				mockTokenVerifier.Claims.Set("sub", fbUser2.FirebaseUID)
				userResource.TokenVerifier = mockTokenVerifier

				result, err = userResource.Update(fbUser, request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) cannot create/update a user associated with another firebase user and 0 more errors, Error updating user, user " + fbUser2.GetID() + " (firebaseUID " + fbUser2.FirebaseUID + ") attempting to create/update user " + fbUser.GetID() + " (firebaseUID " + fbUser.FirebaseUID + ")"))
			})

		})

		Context("update a user not associated with firebase", func() {

			Context("requesting user is car share admin", func() {

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

				It("should return a model.User result payload", func() {
					Expect(result.Result()).To(BeAssignableToTypeOf(model.User{}))
				})

				It("should persist and return the updated user", func() {
					resUser := result.Result().(model.User)
					csLinkedUser.ID = resUser.ID
					Expect(resUser).To(Equal(csLinkedUser))
					persistedUser, err := userResource.UserStorage.GetOne(resUser.GetID(), request.Context)
					Expect(err).ToNot(HaveOccurred())
					Expect(persistedUser).To(Equal(csLinkedUser))
				})

			})

			Context("requesting user not car share admin", func() {

				BeforeEach(func() {

					// fbUser2 is not an admin for the car share csLinkedUser is linked to
					mockTokenVerifier := mockTokenVerifier{}
					mockTokenVerifier.Claims = make(jwt.Claims)
					mockTokenVerifier.Claims.Set("sub", fbUser2.FirebaseUID)
					userResource.TokenVerifier = mockTokenVerifier

					csLinkedUser.DisplayName = csLinkedUser.DisplayName + " updated"
					result, err = userResource.Update(csLinkedUser, request)
				})

				It("should throw an error", func() {
					Expect(err).To(HaveOccurred())
				})
			})

		})

	})

	Describe("delete", func() {

		var (
			result api2go.Responder
			err    error
		)

		Context("firebase user", func() {

			BeforeEach(func() {

				// simulate the request coming in with a valid JWT token for the
				// user being deleted
				mockTokenVerifier := mockTokenVerifier{}
				mockTokenVerifier.Claims = make(jwt.Claims)
				mockTokenVerifier.Claims.Set("sub", fbUser.FirebaseUID)
				userResource.TokenVerifier = mockTokenVerifier

				result, err = userResource.Delete(fbUser.GetID(), request)
			})

			It("should throw an error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) unable to delete users linked to Firebase and 0 more errors, error deleting user, user " + fbUser.GetID() + " attempting to delete firebase user " + fbUser.GetID()))
			})

		})

		Context("user linked to car share", func() {

			Context("requesting user is admin for car share", func() {

				BeforeEach(func() {

					// simulate the request coming in with a valid JWT token for an admin for the car share that the target user is linked to
					mockTokenVerifier := mockTokenVerifier{}
					mockTokenVerifier.Claims = make(jwt.Claims)
					mockTokenVerifier.Claims.Set("sub", fbUser.FirebaseUID)
					userResource.TokenVerifier = mockTokenVerifier

					result, err = userResource.Delete(csLinkedUser.GetID(), request)
				})

				It("should not throw an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should delete the user from the data store", func() {
					_, err = userResource.UserStorage.GetOne(csLinkedUser.GetID(), request.Context)
					Expect(err).To(Equal(storage.ErrNotFound))
				})

				It("should remove the user from the linked car share", func() {
					carShare, err := userResource.CarShareStorage.GetOne(csLinkedUser.LinkedCarShareID, request.Context)
					Expect(err).ToNot(HaveOccurred())
					Expect(carShare.MemberIDs).NotTo(ContainElement(csLinkedUser.GetID()))
				})

			})

			Context("requesting user is not admin for car share", func() {

				BeforeEach(func() {

					// simulate the request coming in with a valid JWT token, but for a user that is not an admin for the car share that the target user is linked to
					mockTokenVerifier := mockTokenVerifier{}
					mockTokenVerifier.Claims = make(jwt.Claims)
					mockTokenVerifier.Claims.Set("sub", fbUser2.FirebaseUID)
					userResource.TokenVerifier = mockTokenVerifier

					result, err = userResource.Delete(csLinkedUser.GetID(), request)
				})

				It("should throw an error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal(
						"http error (403) only admins for car share " + carShare.GetID() + " can delete user " +
							csLinkedUser.GetID() + " and 0 more errors, error deleting user, user " + fbUser2.GetID() +
							" attempting to delete user " + csLinkedUser.GetID() + " linked to car share " + carShare.GetID() +
							", but isn't an admin"))
				})

			})

		})

		Context("user that doesnt exist", func() {

			BeforeEach(func() {
				// simulate the request coming in with a valid JWT token for an admin for the car share that the target user is linked to
				mockTokenVerifier := mockTokenVerifier{}
				mockTokenVerifier.Claims = make(jwt.Claims)
				mockTokenVerifier.Claims.Set("sub", fbUser.FirebaseUID)
				userResource.TokenVerifier = mockTokenVerifier
			})

			Context("\"valid\" id, just user just doesn't exist", func() {

				var target = bson.NewObjectId().Hex()

				BeforeEach(func() {
					result, err = userResource.Delete(target, request)
				})

				It("should return a 400 error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("http error (400) error retrieving target user, not found and 0 more errors, error deleting user, target user " + target + " not found"))
				})

			})

			Context("invalid id", func() {

				var target = "invalid id"

				BeforeEach(func() {
					result, err = userResource.Delete(target, request)
				})

				It("should return a 500 error", func() {
					Expect(err).To(HaveOccurred())
					Expect(err.Error()).To(Equal("http error (500) error retrieving target user, invalid ID and 0 more errors, error deleting user, error retrieving target user " + target + ", invalid ID"))
				})

			})

		})

		Context("requesting user not logged in", func() {

			BeforeEach(func() {
				mockTokenVerifier := userResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Error = fmt.Errorf("example error")
				userResource.TokenVerifier = mockTokenVerifier
				result, err = userResource.Delete(csLinkedUser.GetID(), request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, error deleting user, example error"))
			})

		})

	})

})
