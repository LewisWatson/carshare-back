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
var _ = Describe("CarShareBack", func() {
	var rec *httptest.ResponseRecorder

	BeforeEach(func() {
		api = api2go.NewAPIWithBaseURL("v0", "http://localhost:31415")
		tripStorage := storage.NewTripStorage()
		userStorage := storage.NewUserStorage()
		carShareStorage := storage.NewCarShareStorage()
		scoreStorage := storage.NewScoreStorage()
		api.AddResource(model.User{}, resource.UserResource{UserStorage: userStorage})
		api.AddResource(model.Trip{}, resource.TripResource{TripStorage: tripStorage, UserStorage: userStorage, CarShareStorage: carShareStorage, ScoreStorage: scoreStorage})
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
			}
		}
		`))
	}

	It("Creates a new user", func() {
		createUser()
	})

	var createCarShare = func() {
		rec = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/v0/carShares", strings.NewReader(`
		{
			"data": {
				"type": "carShares",
				"attributes": {
					"name": "carShare1",
					"metres": 1000
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
			    "type": "carShares",
			    "id": "1",
			    "attributes": {
			      "name": "carShare1"
			    },
			    "relationships": {
			      "admins": {
			        "links": {
			          "self": "http://localhost:31415/v0/carShares/1/relationships/admins",
			          "related": "http://localhost:31415/v0/carShares/1/admins"
			        },
			        "data": []
			      },
			      "trips": {
			        "links": {
			          "self": "http://localhost:31415/v0/carShares/1/relationships/trips",
			          "related": "http://localhost:31415/v0/carShares/1/trips"
			        },
			        "data": []
			      }
			    }
			  }
			}
		`))
	}

	It("Creates a new car share", func() {
		createCarShare()
	})

	var createTrip = func() {
		rec = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/v0/trips", strings.NewReader(`
		{
			"data": {
				"type": "trips",
				"attributes": {
					"meters-as-driver": 1000,
					"meters-as-passenger": 1000
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
			    "type": "trips",
			    "id": "1",
			    "attributes": {
			      "metres": 0,
			      "timestamp": "0001-01-01T00:00:00Z"
			    },
			    "relationships": {
			      "carShare": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/carShare",
			          "related": "http://localhost:31415/v0/trips/1/carShare"
			        },
			        "data": null
			      },
			      "driver": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/driver",
			          "related": "http://localhost:31415/v0/trips/1/driver"
			        },
			        "data": null
			      },
			      "passengers": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/passengers",
			          "related": "http://localhost:31415/v0/trips/1/passengers"
			        },
			        "data": []
			      },
			      "scores": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/scores",
			          "related": "http://localhost:31415/v0/trips/1/scores"
			        },
			        "data": []
			      }
			    }
			  }
			}
		`))
	}

	It("Creates a trip", func() {
		createTrip()
	})

	It("Adds a driver to a trip", func() {
		createUser()
		createTrip()
		rec = httptest.NewRecorder()

		By("Adding a driver to a trip with PATCH")

		req, err := http.NewRequest("PATCH", "/v0/trips/1", strings.NewReader(`
		{
		  "data": {
		    "type": "trips",
		    "id": "1",
		    "attributes": {},
		    "relationships": {
		      "driver": {
		        "data": {
		          "type": "users",
		          "id": "1"
		        }
		      }
		    }
		  }
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the trip from the backend, it should have the user")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/trips/1", nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(`
			{
			  "data": {
			    "type": "trips",
			    "id": "1",
			    "attributes": {
			      "metres": 0,
			      "timestamp": "0001-01-01T00:00:00Z"
			    },
			    "relationships": {
			      "carShare": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/carShare",
			          "related": "http://localhost:31415/v0/trips/1/carShare"
			        },
			        "data": null
			      },
			      "driver": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/driver",
			          "related": "http://localhost:31415/v0/trips/1/driver"
			        },
			        "data": {
			          "type": "users",
			          "id": "1"
			        }
			      },
			      "passengers": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/passengers",
			          "related": "http://localhost:31415/v0/trips/1/passengers"
			        },
			        "data": []
			      },
			      "scores": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/scores",
			          "related": "http://localhost:31415/v0/trips/1/scores"
			        },
			        "data": []
			      }
			    }
			  },
			  "included": [
			    {
			      "type": "users",
			      "id": "1",
			      "attributes": {
			        "user-name": "marvin"
			      }
			    }
			  ]
			}
		`))
	})

	It("Adds a car share to a trip", func() {
		createCarShare()
		createTrip()
		rec = httptest.NewRecorder()

		By("Adding a carShare to a trip with PATCH")

		req, err := http.NewRequest("PATCH", "/v0/trips/1", strings.NewReader(`
		{
			"data": {
				"type": "trips",
				"id": "1",
				"attributes": {},
				"relationships": {
					"carShare": {
						"data": {
							"type": "carShares",
							"id": "1"
						}
					}
				}
			}
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the trip from the backend, it should have the car share")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/trips/1", nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(`
			{
			  "data": {
			    "type": "trips",
			    "id": "1",
			    "attributes": {
			      "metres": 0,
			      "timestamp": "0001-01-01T00:00:00Z"
			    },
			    "relationships": {
			      "carShare": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/carShare",
			          "related": "http://localhost:31415/v0/trips/1/carShare"
			        },
			        "data": {
			          "type": "carShares",
			          "id": "1"
			        }
			      },
			      "driver": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/driver",
			          "related": "http://localhost:31415/v0/trips/1/driver"
			        },
			        "data": null
			      },
			      "passengers": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/passengers",
			          "related": "http://localhost:31415/v0/trips/1/passengers"
			        },
			        "data": []
			      },
			      "scores": {
			        "links": {
			          "self": "http://localhost:31415/v0/trips/1/relationships/scores",
			          "related": "http://localhost:31415/v0/trips/1/scores"
			        },
			        "data": []
			      }
			    }
			  },
			  "included": [
			    {
			      "type": "carShares",
			      "id": "1",
			      "attributes": {
			        "name": "carShare1"
			      },
			      "relationships": {
			        "admins": {
			          "links": {
			            "self": "http://localhost:31415/v0/carShares/1/relationships/admins",
			            "related": "http://localhost:31415/v0/carShares/1/admins"
			          },
			          "data": []
			        },
			        "trips": {
			          "links": {
			            "self": "http://localhost:31415/v0/carShares/1/relationships/trips",
			            "related": "http://localhost:31415/v0/carShares/1/trips"
			          },
			          "data": []
			        }
			      }
			    }
			  ]
			}
		`))
	})

	It("Adds a trip to a car share", func() {
		createUser()
		createCarShare()
		createTrip()
		rec = httptest.NewRecorder()

		By("Adding a trip with POST")

		req, err := http.NewRequest("POST", "/v0/carShares/1/relationships/trips", strings.NewReader(`
		{
			"data": [{
				"type": "trips",
				"id": "1"
			}]
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the car share from the backend, it should have the trip")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/carShares/1", nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(`
			{
			  "data": {
			    "type": "carShares",
			    "id": "1",
			    "attributes": {
			      "name": "carShare1"
			    },
			    "relationships": {
			      "admins": {
			        "links": {
			          "self": "http://localhost:31415/v0/carShares/1/relationships/admins",
			          "related": "http://localhost:31415/v0/carShares/1/admins"
			        },
			        "data": []
			      },
			      "trips": {
			        "links": {
			          "self": "http://localhost:31415/v0/carShares/1/relationships/trips",
			          "related": "http://localhost:31415/v0/carShares/1/trips"
			        },
			        "data": [
			          {
			            "type": "trips",
			            "id": "1"
			          }
			        ]
			      }
			    }
			  },
			  "included": [
			    {
			      "type": "trips",
			      "id": "1",
			      "attributes": {
			        "metres": 0,
			        "timestamp": "0001-01-01T00:00:00Z"
			      },
			      "relationships": {
			        "carShare": {
			          "links": {
			            "self": "http://localhost:31415/v0/trips/1/relationships/carShare",
			            "related": "http://localhost:31415/v0/trips/1/carShare"
			          },
			          "data": null
			        },
			        "driver": {
			          "links": {
			            "self": "http://localhost:31415/v0/trips/1/relationships/driver",
			            "related": "http://localhost:31415/v0/trips/1/driver"
			          },
			          "data": null
			        },
			        "passengers": {
			          "links": {
			            "self": "http://localhost:31415/v0/trips/1/relationships/passengers",
			            "related": "http://localhost:31415/v0/trips/1/passengers"
			          },
			          "data": []
			        },
			        "scores": {
			          "links": {
			            "self": "http://localhost:31415/v0/trips/1/relationships/scores",
			            "related": "http://localhost:31415/v0/trips/1/scores"
			          },
			          "data": []
			        }
			      }
			    }
			  ]
			}
		`))
	})

	var replaceTrips = func() {
		rec = httptest.NewRecorder()
		By("Replacing trip relationship with PATCH")

		req, err := http.NewRequest("PATCH", "/v0/carShares/1/relationships/trips", strings.NewReader(`
		{
			"data": [{
				"type": "trips",
				"id": "1"
			}]
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the car share from the backend, it should have the relationship")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/carShares/1", nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(`
			{
			  "data": {
			    "type": "carShares",
			    "id": "1",
			    "attributes": {
			      "name": "carShare1"
			    },
			    "relationships": {
			      "admins": {
			        "links": {
			          "self": "http://localhost:31415/v0/carShares/1/relationships/admins",
			          "related": "http://localhost:31415/v0/carShares/1/admins"
			        },
			        "data": []
			      },
			      "trips": {
			        "links": {
			          "self": "http://localhost:31415/v0/carShares/1/relationships/trips",
			          "related": "http://localhost:31415/v0/carShares/1/trips"
			        },
			        "data": [
			          {
			            "type": "trips",
			            "id": "1"
			          }
			        ]
			      }
			    }
			  },
			  "included": [
			    {
			      "type": "trips",
			      "id": "1",
			      "attributes": {
			        "metres": 0,
			        "timestamp": "0001-01-01T00:00:00Z"
			      },
			      "relationships": {
			        "carShare": {
			          "links": {
			            "self": "http://localhost:31415/v0/trips/1/relationships/carShare",
			            "related": "http://localhost:31415/v0/trips/1/carShare"
			          },
			          "data": null
			        },
			        "driver": {
			          "links": {
			            "self": "http://localhost:31415/v0/trips/1/relationships/driver",
			            "related": "http://localhost:31415/v0/trips/1/driver"
			          },
			          "data": null
			        },
			        "passengers": {
			          "links": {
			            "self": "http://localhost:31415/v0/trips/1/relationships/passengers",
			            "related": "http://localhost:31415/v0/trips/1/passengers"
			          },
			          "data": []
			        },
			        "scores": {
			          "links": {
			            "self": "http://localhost:31415/v0/trips/1/relationships/scores",
			            "related": "http://localhost:31415/v0/trips/1/scores"
			          },
			          "data": []
			        }
			      }
			    }
			  ]
			}
		`))
	}

	It("Replaces car share's trips", func() {
		createUser()
		createCarShare()
		createTrip()
		replaceTrips()
	})

	It("Deletes a car share trip", func() {
		createUser()
		createCarShare()
		createTrip()
		replaceTrips()
		rec = httptest.NewRecorder()

		By("Deleting the car shares only trip with ID 1")

		req, err := http.NewRequest("DELETE", "/v0/carShares/1/relationships/trips", strings.NewReader(`
		{
			"data": [{
				"type": "trips",
				"id": "1"
			}]
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the car share from the backend, it should not have any relations")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/carShares/1", nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(`
			{
			  "data": {
			    "type": "carShares",
			    "id": "1",
			    "attributes": {
			      "name": "carShare1"
			    },
			    "relationships": {
			      "admins": {
			        "links": {
			          "self": "http://localhost:31415/v0/carShares/1/relationships/admins",
			          "related": "http://localhost:31415/v0/carShares/1/admins"
			        },
			        "data": []
			      },
			      "trips": {
			        "links": {
			          "self": "http://localhost:31415/v0/carShares/1/relationships/trips",
			          "related": "http://localhost:31415/v0/carShares/1/trips"
			        },
			        "data": []
			      }
			    }
			  }
			}
		`))
	})
})
