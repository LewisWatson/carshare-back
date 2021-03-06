package resource

import (
	"fmt"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/LewisWatson/carshare-back/storage/mongodb"
	"github.com/manyminds/api2go"
	
	"gopkg.in/jose.v1/jwt"
	"gopkg.in/mgo.v2/bson"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("car share resource", func() {

	var (
		carShareResource *CarShareResource
		request          api2go.Request
		context          *api2go.APIContext

		user1ID     = bson.NewObjectId()
		user2ID     = bson.NewObjectId()
		user3ID     = bson.NewObjectId()
		carShare1ID = bson.NewObjectId()
		carShare2ID = bson.NewObjectId()
		trip1ID     = bson.NewObjectId()
		trip2ID     = bson.NewObjectId()
		trip3ID     = bson.NewObjectId()
	)

	BeforeEach(func() {
		mockTokenVerifier := mockTokenVerifier{}
		mockTokenVerifier.Claims = make(jwt.Claims)
		mockTokenVerifier.Claims.Set("sub", "user1FirebaseUID")
		carShareResource = &CarShareResource{
			CarShareStorage: &mongodb.CarShareStorage{},
			TripStorage:     &mongodb.TripStorage{},
			UserStorage:     &mongodb.UserStorage{},
			TokenVerifier:   mockTokenVerifier,
		}
		context = &api2go.APIContext{}
		db, pool, containerResource = mongodb.ConnectToMongoDB(db, pool, containerResource)
		Expect(db).ToNot(BeNil())
		Expect(pool).ToNot(BeNil())
		Expect(containerResource).ToNot(BeNil())
		err := db.DB(mongodb.CarShareDB).DropDatabase()
		Expect(err).ToNot(HaveOccurred())
		context.Set("db", db)
		request = api2go.Request{Context: context}
		db.DB(mongodb.CarShareDB).C(mongodb.UsersColl).Insert(
			&model.User{ID: user1ID, FirebaseUID: "user1FirebaseUID"},
			&model.User{ID: user2ID, FirebaseUID: "user2FirebaseUID"},
			&model.User{ID: user3ID, FirebaseUID: "user3FirebaseUID"},
		)
		db.DB(mongodb.CarShareDB).C(mongodb.CarSharesColl).Insert(
			&model.CarShare{
				ID:        carShare1ID,
				TripIDs:   []string{trip1ID.Hex()},
				MemberIDs: []string{user1ID.Hex()},
				AdminIDs:  []string{user1ID.Hex()},
			},
			&model.CarShare{
				ID:       carShare2ID,
				AdminIDs: []string{user1ID.Hex()},
			},
		)
		db.DB(mongodb.CarShareDB).C(mongodb.TripsColl).Insert(
			&model.Trip{
				ID:         trip1ID,
				Metres:     123,
				CarShareID: carShare1ID.Hex(),
			},
			&model.Trip{
				ID:     trip2ID,
				Metres: 456,
			},
			&model.Trip{
				ID:           trip3ID,
				Metres:       789,
				DriverID:     user1ID.Hex(),
				PassengerIDs: []string{user2ID.Hex(), user3ID.Hex()},
			},
		)
	})

	Describe("get all", func() {

		var (
			carShares []model.CarShare
			result    api2go.Responder
			err       error
		)

		BeforeEach(func() {
			carShares, err = carShareResource.CarShareStorage.GetAll(user1ID.Hex(), context)
			Expect(err).NotTo(HaveOccurred())
			Expect(carShares).To(HaveLen(1))
			result, err = carShareResource.FindAll(request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return car shares that the user is a member of", func() {
			Expect(result).ToNot(BeNil())
			response, ok := result.(*Response)
			Expect(ok).To(BeTrue())
			Expect(response.Res).To(Equal(carShares))
		})

		Context("user not logged in", func() {

			BeforeEach(func() {
				mockTokenVerifier := carShareResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Error = fmt.Errorf("example error")
				carShareResource.TokenVerifier = mockTokenVerifier
				result, err = carShareResource.FindAll(request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, example error"))
			})

		})

	})

	Describe("get one", func() {

		var (
			carShare model.CarShare
			result   api2go.Responder
			err      error
		)

		BeforeEach(func() {
			carShare, err = carShareResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
			Expect(err).NotTo(HaveOccurred())
			Expect(carShare).NotTo(BeNil())
			result, err = carShareResource.FindOne(carShare1ID.Hex(), request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return target car share", func() {
			Expect(result).ToNot(BeNil())
			response, ok := result.(*Response)
			Expect(ok).To(BeTrue())
			Expect(response.Res).To(BeAssignableToTypeOf(model.CarShare{}))
			responseCarShare := response.Res.(model.CarShare)
			Expect(responseCarShare.GetID()).To(Equal(carShare.GetID()))
			Expect(responseCarShare.TripIDs).To(Equal(carShare.TripIDs))
			Expect(responseCarShare.MemberIDs).To(Equal(carShare.MemberIDs))
			Expect(responseCarShare.AdminIDs).To(Equal(carShare.AdminIDs))
		})

		Context("invalid id", func() {

			Context("trip does not exist", func() {

				var carShareID = bson.NewObjectId().Hex()

				BeforeEach(func() {
					result, err = carShareResource.FindOne(carShareID, request)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

				It("should return an api2go.HTTPError", func() {
					Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
				})

				It("should return an api2go.HTTPError with the correct message", func() {
					actual := err.(api2go.HTTPError).Error()
					expected := fmt.Sprintf("http error (404) Not Found and 0 more errors, unable to find car share %s", carShareID)
					Expect(actual).To(Equal(expected))
				})

			})

			Context("invalid bson ID", func() {

				var carShareID = "invalid"

				BeforeEach(func() {
					result, err = carShareResource.FindOne(carShareID, request)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

				It("should return an api2go.HTTPError", func() {
					Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
				})

				It("should return an api2go.HTTPError with the correct message", func() {
					actual := err.(api2go.HTTPError).Error()
					expected := fmt.Sprintf("http error (500) Error occurred while retrieving car share %s and 0 more errors, Error occurred while retrieving car share %s, invalid ID", carShareID, carShareID)
					Expect(actual).To(Equal(expected))
				})

			})

		})

		Context("user not logged in", func() {

			BeforeEach(func() {
				mockTokenVerifier := carShareResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Error = fmt.Errorf("example error")
				carShareResource.TokenVerifier = mockTokenVerifier
				result, err = carShareResource.FindOne("example id", request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, Error retrieving car share, example error"))
			})

		})

	})

	Describe("create", func() {

		var (
			carshare = model.CarShare{
				Name: "example car share",
			}
			result api2go.Responder
			err    error
		)

		BeforeEach(func() {
			result, err = carShareResource.Create(carshare, request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return http status created", func() {
			Expect(result.StatusCode()).To(Equal(http.StatusCreated))
		})

		It("should persist and return the car share with user as admin", func() {

			By("Add user1 as member and admin")
			user, err := carShareResource.UserStorage.GetOne(user1ID.Hex(), context)
			Expect(err).ToNot(HaveOccurred())
			carshare.AdminIDs = append(carshare.AdminIDs, user.GetID())
			carshare.MemberIDs = append(carshare.MemberIDs, user.GetID())

			By("return populated car share in response")
			carshare.Members = append(carshare.Members, &user)
			carshare.Admins = append(carshare.Admins, &user)
			Expect(result.Result()).To(BeAssignableToTypeOf(model.CarShare{}))
			resCarshare := result.Result().(model.CarShare)
			carshare.ID = resCarshare.ID
			Expect(resCarshare).To(Equal(carshare))

			By("persist car share in data store")
			carshare.Admins = nil
			carshare.Members = nil
			carshare.TripIDs = []string{}
			persistedCarShare, err := carShareResource.CarShareStorage.GetOne(resCarshare.GetID(), context)
			Expect(err).ToNot(HaveOccurred())
			Expect(persistedCarShare).To(Equal(carshare))
		})

		Context("user not logged in", func() {

			BeforeEach(func() {
				mockTokenVerifier := carShareResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Error = fmt.Errorf("example error")
				carShareResource.TokenVerifier = mockTokenVerifier
				result, err = carShareResource.Create(carshare, request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, Error creating car share, example error"))
			})

		})

	})

	Describe("update", func() {

		var (
			carShare model.CarShare
			result   api2go.Responder
			err      error
		)

		Context("attribute", func() {

			BeforeEach(func() {
				carShare, err = carShareResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
				Expect(err).NotTo(HaveOccurred())
				Expect(carShare).NotTo(BeNil())
				carShare.Name = "updated"
				result, err = carShareResource.Update(carShare, request)
			})

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return updated target car share", func() {
				Expect(result).ToNot(BeNil())
				response, ok := result.(*Response)
				Expect(ok).To(Equal(true))
				Expect(response.Res).To(BeAssignableToTypeOf(model.CarShare{}))
				responseCarShare := response.Res.(model.CarShare)
				Expect(responseCarShare.GetID()).To(Equal(carShare.GetID()))
				Expect(responseCarShare.Name).To(Equal("updated"))
			})

			It("should update carShare in data store", func() {
				carShare, err = carShareResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
				Expect(err).NotTo(HaveOccurred())
				Expect(carShare).NotTo(BeNil())
				Expect(carShare.Name).To(Equal("updated"))
			})

		})

		Context("relationship", func() {

			BeforeEach(func() {
				carShare, err = carShareResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
				Expect(err).NotTo(HaveOccurred())
				Expect(carShare).NotTo(BeNil())
			})

			Context("hasMany trips", func() {

				Context("valid trips", func() {

					BeforeEach(func() {
						carShare.TripIDs = append(carShare.TripIDs, trip2ID.Hex(), trip3ID.Hex())
						result, err = carShareResource.Update(carShare, request)
					})

					It("should not throw an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					It("should return updated target car share", func() {
						Expect(result).ToNot(BeNil())
						response, ok := result.(*Response)
						Expect(ok).To(BeTrue())
						Expect(response.Res).To(BeAssignableToTypeOf(model.CarShare{}))
						resCarShare := response.Res.(model.CarShare)
						Expect(resCarShare.GetID()).To(Equal(carShare.GetID()))
						Expect(resCarShare.TripIDs).To(ConsistOf(trip1ID.Hex(), trip2ID.Hex(), trip3ID.Hex()))
					})

					Specify("target car share should have the trips in the data store", func() {
						carShare, err = carShareResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(carShare).NotTo(BeNil())
						Expect(carShare.TripIDs).To(ConsistOf(trip1ID.Hex(), trip2ID.Hex(), trip3ID.Hex()))
					})

					Specify("trip 1 should belong to car share 1", func() {
						trip, err := carShareResource.TripStorage.GetOne(trip1ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(trip.CarShareID).To(Equal(carShare1ID.Hex()))
					})

					Specify("trip 2 should belong to car share 1", func() {
						trip, err := carShareResource.TripStorage.GetOne(trip2ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(trip.CarShareID).To(Equal(carShare1ID.Hex()))
					})

					Specify("trip 3 should belong to car share 1", func() {
						trip, err := carShareResource.TripStorage.GetOne(trip3ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(trip.CarShareID).To(Equal(carShare1ID.Hex()))
					})
				})

				Context("invalid trips", func() {

					Context("trip doesnt exist", func() {

						var dodgyTripID = bson.NewObjectId().Hex()

						BeforeEach(func() {
							carShare.TripIDs = append(carShare.TripIDs, dodgyTripID)
							result, err = carShareResource.Update(carShare, request)
						})

						It("should throw an error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
							expectedErr := fmt.Sprintf("Error verifying trip %s", dodgyTripID)
							expectedHTTPErr := api2go.NewHTTPError(
								fmt.Errorf("%s, %s", expectedErr, storage.ErrNotFound),
								expectedErr,
								http.StatusInternalServerError,
							)
							Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
						})

					})

					Context("trip already belongs to another car share", func() {

						BeforeEach(func() {
							carShare, err = carShareResource.CarShareStorage.GetOne(carShare2ID.Hex(), context)
							Expect(err).NotTo(HaveOccurred())
							Expect(carShare).NotTo(BeNil())
							carShare.TripIDs = append(carShare.TripIDs, trip1ID.Hex())
							result, err = carShareResource.Update(carShare, request)
						})

						It("should throw an error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
							expectedErr := fmt.Sprintf("trip %s already belongs to another car share", trip1ID.Hex())
							expectedHTTPErr := api2go.NewHTTPError(
								fmt.Errorf("%s", expectedErr),
								expectedErr,
								http.StatusInternalServerError,
							)
							Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
						})

					})

				})

			})

			Context("hasMany members", func() {

				Context("valid member", func() {

					BeforeEach(func() {
						carShare.MemberIDs = append(carShare.MemberIDs, user2ID.Hex(), user3ID.Hex())
						result, err = carShareResource.Update(carShare, request)
					})

					It("should not throw an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					It("should return updated target car share", func() {
						Expect(result).ToNot(BeNil())
						response, ok := result.(*Response)
						Expect(ok).To(BeTrue())
						Expect(response.Res).To(BeAssignableToTypeOf(model.CarShare{}))
						resCarShare := response.Res.(model.CarShare)
						Expect(resCarShare.GetID()).To(Equal(carShare.GetID()))
						Expect(resCarShare.MemberIDs).To(ConsistOf(user1ID.Hex(), user2ID.Hex(), user3ID.Hex()))
					})

					Specify("target car share should have the members in the data store", func() {
						carShare, err = carShareResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(carShare).NotTo(BeNil())
						Expect(carShare.MemberIDs).To(ConsistOf(user1ID.Hex(), user2ID.Hex(), user3ID.Hex()))
					})
				})

				Context("invalid member", func() {

					Context("user doesnt exist", func() {

						var dodgyUserID = bson.NewObjectId().Hex()

						BeforeEach(func() {
							carShare.AdminIDs = append(carShare.MemberIDs, dodgyUserID)
							result, err = carShareResource.Update(carShare, request)
						})

						It("should throw an error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
							expectedErr := fmt.Sprintf("Error verifying user %s", dodgyUserID)
							expectedHTTPErr := api2go.NewHTTPError(
								fmt.Errorf("%s, %s", expectedErr, storage.ErrNotFound),
								expectedErr,
								http.StatusInternalServerError,
							)
							Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
						})

					})

				})

			})

			Context("hasMany admins", func() {

				Context("valid admin", func() {

					BeforeEach(func() {
						carShare.AdminIDs = append(carShare.AdminIDs, user2ID.Hex(), user3ID.Hex())
						result, err = carShareResource.Update(carShare, request)
					})

					It("should not throw an error", func() {
						Expect(err).ToNot(HaveOccurred())
					})

					It("should return updated target car share", func() {
						Expect(result).ToNot(BeNil())
						response, ok := result.(*Response)
						Expect(ok).To(BeTrue())
						Expect(response.Res).To(BeAssignableToTypeOf(model.CarShare{}))
						resCarShare := response.Res.(model.CarShare)
						Expect(resCarShare.GetID()).To(Equal(carShare.GetID()))
						Expect(resCarShare.AdminIDs).To(ConsistOf(user1ID.Hex(), user2ID.Hex(), user3ID.Hex()))
					})

					Specify("target car share should have the admins in the data store", func() {
						carShare, err = carShareResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(carShare).NotTo(BeNil())
						Expect(carShare.AdminIDs).To(ConsistOf(user1ID.Hex(), user2ID.Hex(), user3ID.Hex()))
					})
				})

				Context("invalid admins", func() {

					Context("user doesnt exist", func() {

						var dodgyUserID = bson.NewObjectId().Hex()

						BeforeEach(func() {
							carShare.AdminIDs = append(carShare.AdminIDs, dodgyUserID)
							result, err = carShareResource.Update(carShare, request)
						})

						It("should throw an error", func() {
							Expect(err).To(HaveOccurred())
							Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
							expectedErr := fmt.Sprintf("Error verifying user %s", dodgyUserID)
							expectedHTTPErr := api2go.NewHTTPError(
								fmt.Errorf("%s, %s", expectedErr, storage.ErrNotFound),
								expectedErr,
								http.StatusInternalServerError,
							)
							Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
						})

					})

				})

			})

		})

		Context("non-admin user attempting update", func() {

			BeforeEach(func() {
				mockTokenVerifier := carShareResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Claims.Set("sub", "user2FirebaseUID")
				carShareResource.TokenVerifier = mockTokenVerifier
				result, err = carShareResource.Update(carShare, request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, Non admin user " + user2ID.Hex() + " attempting to update carShare " + carShare.GetID()))
			})

		})

		Context("user not logged in", func() {

			BeforeEach(func() {
				mockTokenVerifier := carShareResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Error = fmt.Errorf("example error")
				carShareResource.TokenVerifier = mockTokenVerifier
				result, err = carShareResource.Update(carShare, request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, Error updating car share, example error"))
			})

		})

	})

	Describe("delete", func() {

		var (
			carShare model.CarShare
			result   api2go.Responder
			err      error
		)

		BeforeEach(func() {
			result, err = carShareResource.Delete(carShare1ID.Hex(), request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return http status ok", func() {
			Expect(result).ToNot(BeNil())
			response, ok := result.(*Response)
			Expect(ok).To(Equal(true))
			Expect(response.Code).To(Equal(http.StatusOK))
		})

		It("car share should be deleted from data store", func() {
			carShare, err = carShareResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(storage.ErrNotFound))
		})

		It("should also delete trips that belong to the car share", func() {
			_, err := carShareResource.TripStorage.GetOne(trip1ID.Hex(), context)
			Expect(err).To(HaveOccurred())
			Expect(err).To(Equal(storage.ErrNotFound))
		})

		Context("invalid id", func() {

			Context("trip does not exist", func() {

				var carShareID = bson.NewObjectId().Hex()

				BeforeEach(func() {
					result, err = carShareResource.Delete(carShareID, request)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

				It("should return an api2go.HTTPError", func() {
					Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
				})

				It("should return an api2go.HTTPError with the correct message", func() {
					actual := err.(api2go.HTTPError).Error()
					expected := fmt.Sprintf("http error (404) Not Found and 0 more errors, unable to find car share %s", carShareID)
					Expect(actual).To(Equal(expected))
				})

			})

			Context("invalid bson ID", func() {

				var carShareID = "invalid"

				BeforeEach(func() {
					result, err = carShareResource.Delete(carShareID, request)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

				It("should return an api2go.HTTPError", func() {
					Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
				})

				It("should return an api2go.HTTPError with the correct message", func() {
					actual := err.(api2go.HTTPError).Error()
					expected := fmt.Sprintf("http error (500) Error occurred while retrieving car share %s and 0 more errors, Error occurred while retrieving car share %s, invalid ID", carShareID, carShareID)
					Expect(actual).To(Equal(expected))
				})

			})

		})

		Context("non-admin user attempting delete", func() {

			BeforeEach(func() {
				mockTokenVerifier := carShareResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Claims.Set("sub", "user2FirebaseUID")
				carShareResource.TokenVerifier = mockTokenVerifier
				result, err = carShareResource.Delete(carShare2ID.Hex(), request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, Non admin user " + user2ID.Hex() + " attempting to delete car share " + carShare2ID.Hex()))
			})

		})

		Context("user not logged in", func() {

			BeforeEach(func() {
				mockTokenVerifier := carShareResource.TokenVerifier.(mockTokenVerifier)
				mockTokenVerifier.Error = fmt.Errorf("example error")
				carShareResource.TokenVerifier = mockTokenVerifier
				result, err = carShareResource.Delete(carShare1ID.Hex(), request)
			})

			It("should return a 403 error", func() {
				Expect(err).To(HaveOccurred())
				Expect(err.Error()).To(Equal("http error (403) Forbidden and 0 more errors, Error deleting car share, example error"))
			})

		})

	})

})
