package resource

import (
	"fmt"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/LewisWatson/carshare-back/storage/mongodb"

	"github.com/manyminds/api2go"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/mgo.v2/bson"
)

var _ = Describe("Trip Resource", func() {

	var (
		tripResource *TripResource
		request      api2go.Request
		context      *api2go.APIContext

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
		tripResource = &TripResource{
			TripStorage:     &mongodb.TripStorage{},
			UserStorage:     &mongodb.UserStorage{},
			CarShareStorage: &mongodb.CarShareStorage{},
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
				ID: user1ID,
			},
			&model.User{
				ID: user2ID,
			},
			&model.User{
				ID: user3ID,
			},
		)
		db.DB(mongodb.CarShareDB).C(mongodb.CarSharesColl).Insert(
			&model.CarShare{
				ID: carShare1ID,
				TripIDs: []string{
					trip1ID.Hex(),
				},
			},
			&model.CarShare{
				ID: carShare2ID,
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
				ID:       trip3ID,
				Metres:   789,
				DriverID: user1ID.Hex(),
				PassengerIDs: []string{
					user2ID.Hex(),
					user3ID.Hex(),
				},
			},
		)
	})

	Describe("get all", func() {

		var (
			trips  []model.Trip
			result api2go.Responder
			err    error
		)

		BeforeEach(func() {
			trips, err = tripResource.TripStorage.GetAll(context)
			Expect(err).NotTo(HaveOccurred())
			Expect(trips).NotTo(BeNil())
			result, err = tripResource.FindAll(request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		// TODO revisit as the trips get popoulated
		// It("should return all existing trips", func() {
		// 	Expect(result).ToNot(BeNil())
		// 	response, ok := result.(*Response)
		// 	Expect(ok).To(Equal(true))
		// 	Expect(response.Res).To(Equal(trips))
		// })

	})

	Describe("get one", func() {

		var (
			trip   model.Trip
			result api2go.Responder
			err    error
		)

		BeforeEach(func() {
			trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
			Expect(err).NotTo(HaveOccurred())
			Expect(trip).NotTo(BeNil())
			result, err = tripResource.FindOne(trip1ID.Hex(), request)
		})

		It("should not throw an error", func() {
			Expect(err).ToNot(HaveOccurred())
		})

		It("should return target trip", func() {
			Expect(result).ToNot(BeNil())
			response, ok := result.(*Response)
			Expect(ok).To(Equal(true))
			Expect(response.Res).To(Equal(trip))
		})

		Context("invalid id", func() {

			Context("trip does not exist", func() {

				var tripID = bson.NewObjectId().Hex()

				BeforeEach(func() {
					result, err = tripResource.FindOne(tripID, request)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

				It("should return an api2go.HTTPError", func() {
					Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
				})

				It("should return an api2go.HTTPError with the correct message", func() {
					actual := err.(api2go.HTTPError).Error()
					expected := fmt.Sprintf("http error (404) Not Found and 0 more errors, unable to find trip %s", tripID)
					Expect(actual).To(Equal(expected))
				})

			})

			Context("invalid bson ID", func() {

				var tripID = "invalid"

				BeforeEach(func() {
					result, err = tripResource.FindOne(tripID, request)
				})

				It("should return an error", func() {
					Expect(err).To(HaveOccurred())
				})

				It("should return an api2go.HTTPError", func() {
					Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
				})

				It("should return an api2go.HTTPError with the correct message", func() {
					actual := err.(api2go.HTTPError).Error()
					expected := fmt.Sprintf("http error (500) Error occurred while retrieving trip %s and 0 more errors, Error occurred while retrieving trip %s, invalid ID", tripID, tripID)
					Expect(actual).To(Equal(expected))
				})

			})

		})

	})

	Describe("update", func() {

		var (
			trip   model.Trip
			result api2go.Responder
			err    error
		)

		Context("update attribute", func() {

			BeforeEach(func() {
				trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
				Expect(err).NotTo(HaveOccurred())
				Expect(trip).NotTo(BeNil())
				trip.Metres = 1337
				result, err = tripResource.Update(trip, request)
			})

			It("should not throw an error", func() {
				Expect(err).ToNot(HaveOccurred())
			})

			It("should return target trip", func() {
				Expect(result).ToNot(BeNil())
				response, ok := result.(*Response)
				Expect(ok).To(Equal(true))
				Expect(response.Res).To(Equal(trip))
			})

			It("should update trip in data store", func() {
				trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
				Expect(err).NotTo(HaveOccurred())
				Expect(trip).NotTo(BeNil())
				Expect(trip.Metres).To(Equal(1337))
			})

		})

		Context("update relationship", func() {

			Context("hasOne car share", func() {

				BeforeEach(func() {
					trip, err = tripResource.TripStorage.GetOne(trip2ID.Hex(), context)
					Expect(err).NotTo(HaveOccurred())
					Expect(trip).NotTo(BeNil())
					trip.CarShareID = carShare1ID.Hex()
					result, err = tripResource.Update(trip, request)
				})

				It("should not throw an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should return target trip", func() {
					Expect(result).ToNot(BeNil())
					response, ok := result.(*Response)
					Expect(ok).To(Equal(true))
					Expect(response.Res).To(Equal(trip))
				})

				Specify("trip should belong to car share", func() {
					trip, err = tripResource.TripStorage.GetOne(trip2ID.Hex(), context)
					Expect(err).NotTo(HaveOccurred())
					Expect(trip).NotTo(BeNil())
					Expect(trip.CarShareID).To(Equal(carShare1ID.Hex()))
				})

				Specify("car share should have trip in list of trips", func() {
					carShare, err := tripResource.CarShareStorage.GetOne(carShare1ID.Hex(), context)
					Expect(err).NotTo(HaveOccurred())
					Expect(trip).NotTo(BeNil())
					Expect(carShare.TripIDs).To(ContainElement(trip2ID.Hex()))
				})

				Context("attempt to re-assign a trip to a different car share", func() {

					BeforeEach(func() {
						trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(trip).NotTo(BeNil())
						trip.CarShareID = carShare2ID.Hex()
						result, err = tripResource.Update(trip, request)
					})

					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
						expectedErr := fmt.Errorf("trip %s already belongs to another car share", trip1ID.Hex())
						expectedHTTPErr := api2go.NewHTTPError(
							expectedErr,
							expectedErr.Error(),
							http.StatusInternalServerError,
						)
						Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
					})

				})

				Context("attempt to re-assign a trip to a car share that doesnt exist", func() {

					BeforeEach(func() {
						trip, err = tripResource.TripStorage.GetOne(trip3ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(trip).NotTo(BeNil())
						trip.CarShareID = bson.NewObjectId().Hex()
						result, err = tripResource.Update(trip, request)
					})

					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
						expectedErr := fmt.Errorf("Unable to find car share %s to in order to add trip relationship", trip.CarShareID)
						expectedHTTPErr := api2go.NewHTTPError(
							expectedErr,
							expectedErr.Error(),
							http.StatusInternalServerError,
						)
						Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
					})

				})

			})

			Context("hasOne driver", func() {

				BeforeEach(func() {
					trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
					Expect(err).NotTo(HaveOccurred())
					Expect(trip).NotTo(BeNil())
					trip.DriverID = user1ID.Hex()
					result, err = tripResource.Update(trip, request)
				})

				It("should not throw an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should return updated target trip", func() {
					Expect(result).ToNot(BeNil())
					response, ok := result.(*Response)
					Expect(ok).To(BeTrue())
					resTrip, ok := response.Res.(model.Trip)
					Expect(ok).To(BeTrue())
					Expect(resTrip.GetID()).To(Equal(trip1ID.Hex()))
					Expect(resTrip.DriverID).To(Equal(user1ID.Hex()))
				})

				Specify("trip should have the driver in data store", func() {
					trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
					Expect(err).NotTo(HaveOccurred())
					Expect(trip).NotTo(BeNil())
					Expect(trip.DriverID).To(Equal(user1ID.Hex()))
				})

				Context("driver doesnt exist", func() {

					BeforeEach(func() {
						trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(trip).NotTo(BeNil())
						trip.DriverID = bson.NewObjectId().Hex()
						result, err = tripResource.Update(trip, request)
					})

					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
						expectedErr := fmt.Sprintf("Error verifying driver %s", trip.DriverID)
						expectedHTTPErr := api2go.NewHTTPError(
							fmt.Errorf("%s, %s", expectedErr, storage.ErrNotFound),
							expectedErr,
							http.StatusInternalServerError,
						)
						Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
					})

				})

			})

			Context("hasMany passengers", func() {

				BeforeEach(func() {
					trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
					Expect(err).NotTo(HaveOccurred())
					Expect(trip).NotTo(BeNil())
					trip.PassengerIDs = append(trip.PassengerIDs, user1ID.Hex(), user2ID.Hex())
					result, err = tripResource.Update(trip, request)
				})

				It("should not throw an error", func() {
					Expect(err).ToNot(HaveOccurred())
				})

				It("should return updated target trip", func() {
					Expect(result).ToNot(BeNil())
					response, ok := result.(*Response)
					Expect(ok).To(BeTrue())
					resTrip, ok := response.Res.(model.Trip)
					Expect(ok).To(BeTrue())
					Expect(resTrip.GetID()).To(Equal(trip1ID.Hex()))
					Expect(resTrip.PassengerIDs).To(ConsistOf(user1ID.Hex(), user2ID.Hex()))
				})

				Specify("trip should have the passengers in data store", func() {
					trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
					Expect(err).NotTo(HaveOccurred())
					Expect(trip).NotTo(BeNil())
					Expect(trip.PassengerIDs).To(ConsistOf(user1ID.Hex(), user2ID.Hex()))
				})

				Context("passenger doesnt exist", func() {

					var dodgyPassengerID = bson.NewObjectId().Hex()

					BeforeEach(func() {
						trip, err = tripResource.TripStorage.GetOne(trip1ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(trip).NotTo(BeNil())
						trip.PassengerIDs = append(trip.PassengerIDs, dodgyPassengerID)
						result, err = tripResource.Update(trip, request)
					})

					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
						expectedErr := fmt.Sprintf("Error verifying passenger %s", dodgyPassengerID)
						expectedHTTPErr := api2go.NewHTTPError(
							fmt.Errorf("%s, %s", expectedErr, storage.ErrNotFound),
							expectedErr,
							http.StatusInternalServerError,
						)
						Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
					})

				})

				Context("passenger already set as driver", func() {

					BeforeEach(func() {
						trip, err = tripResource.TripStorage.GetOne(trip3ID.Hex(), context)
						Expect(err).NotTo(HaveOccurred())
						Expect(trip).NotTo(BeNil())
						trip.PassengerIDs = append(trip.PassengerIDs, trip.DriverID)
						result, err = tripResource.Update(trip, request)
					})

					It("should throw an error", func() {
						Expect(err).To(HaveOccurred())
						Expect(err).To(BeAssignableToTypeOf(api2go.HTTPError{}))
						expectedErr := fmt.Errorf("Error passenger %s is set as driver", trip.DriverID)
						expectedHTTPErr := api2go.NewHTTPError(
							expectedErr,
							expectedErr.Error(),
							http.StatusBadRequest,
						)
						Expect(err.(api2go.HTTPError)).To(Equal(expectedHTTPErr))
					})

				})

			})
		})

	})

})
