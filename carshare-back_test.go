package main_test

import (
	"net/http"
	"net/http/httptest"
	"strings"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/manyminds/api2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

// there are a lot of functions because each test can be run individually and sets up the complete
// environment. That is because we run all the specs randomized.
var _ = Describe("CrudExample", func() {
	var rec *httptest.ResponseRecorder

	BeforeEach(func() {
		api = api2go.NewAPIWithBaseURL("v0", "http://localhost:31415")
		tripStorage := storage.NewTripStorage()
		userStorage := storage.NewUserStorage()
		carShareStorage := storage.NewCarShareStorage()
		api.AddResource(model.User{}, resource.UserResource{UserStorage: userStorage})
		api.AddResource(model.Trip{}, resource.TripResource{TripStorage: tripStorage})
		api.AddResource(model.CarShare{}, resource.CarShareResource{CarShareStorage: carShareStorage, TripStorage: tripStorage, UserStorage: userStorage})
		rec = httptest.NewRecorder()
	})

	var createUser = func() {
		rec = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/v0/users", strings.NewReader(`
		{
			"data": {
				"type": "users",
				"attributes": {
					"user-name": "marvin"
				}
			}
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusCreated))
		Expect(rec.Body.String()).To(MatchJSON(`
    {
      "data": {
        "type": "users",
        "id": "1",
        "attributes": {
          "user-name": "marvin"
        }
      },
      "meta": {
        "author": "Lewis Watson"
      }
    }
		`))
	}

	It("Creates a new user", func() {
		createUser()
	})
})
