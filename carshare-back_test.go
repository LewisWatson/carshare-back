package main_test

import (
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"time"

	mgo "gopkg.in/mgo.v2"
	dockertest "gopkg.in/ory-am/dockertest.v2"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage/in-memory"
	"github.com/LewisWatson/carshare-back/storage/mongodb"
	"github.com/benbjohnson/clock"
	"github.com/manyminds/api2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var db *mgo.Session
var containerID dockertest.ContainerID

/*var _ = BeforeSuite(func() {

	log.Println("Spinning up and connecting to MongoDB container")

	var err error
	containerID, err = dockertest.ConnectToMongoDB(15, time.Millisecond*500, func(url string) bool {
		// This callback function checks if the image's process is responsive.
		// Sometimes, docker images are booted but the process (in this case MongoDB) is still doing maintenance
		// before being fully responsive which might cause issues like "TCP Connection reset by peer".
		var err error
		db, err = mgo.Dial(url)
		if err != nil {
			return false
		}

		// Sometimes, dialing the database is not enough because the port is already open but the process is not responsive.
		// Most database conenctors implement a ping function which can be used to test if the process is responsive.
		// Alternatively, you could execute a query to see if an error occurs or not.
		return db.Ping() == nil
	})

	if err != nil {
		db.Close()
		containerID.KillRemove()
		log.Fatalf("Could not connect to database: %s", err)
	}

	log.Println("Connection to MongoDB established")
})*/

var _ = AfterSuite(func() {

	if db != nil {
		log.Println("Closing connection to MongoDB")
		db.Close()
	}

	if containerID != "" {
		log.Println("Cleaning up MongoDB container")
		containerID.KillRemove()
	}
})

// there are a lot of functions because each test can be run individually and sets up the complete
// environment. That is because we run all the specs randomized.
var _ = Describe("The CarShareBack API", func() {

	var (
		rec       *httptest.ResponseRecorder
		mockClock *clock.Mock
	)

	var createUser = func() string {
		rec := httptest.NewRecorder()
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
		actual := rec.Body.String()
		id := extractIDFromResponse(actual)
		expected := strings.NewReplacer("<<id>>", id).Replace(`
		{
			"data": {
				"type": "users",
				"id": "<<id>>",
				"attributes": {
					"user-name": "marvin"
				}
			}
		}
		`)
		Expect(actual).To(MatchJSON(expected))
		return id
	}

	Describe("Using MongoDB data store", func() {

		var connectToMongoDB = func() {

			if db != nil {
				return
			}

			log.Println("Spinning up and connecting to MongoDB container")

			var err error
			containerID, err = dockertest.ConnectToMongoDB(15, time.Millisecond*500, func(url string) bool {
				// This callback function checks if the image's process is responsive.
				// Sometimes, docker images are booted but the process (in this case MongoDB) is still doing maintenance
				// before being fully responsive which might cause issues like "TCP Connection reset by peer".
				var err error
				db, err = mgo.Dial(url)
				if err != nil {
					return false
				}

				// Sometimes, dialing the database is not enough because the port is already open but the process is not responsive.
				// Most database conenctors implement a ping function which can be used to test if the process is responsive.
				// Alternatively, you could execute a query to see if an error occurs or not.
				return db.Ping() == nil
			})

			if err != nil {
				db.Close()
				containerID.KillRemove()
				log.Fatalf("Could not connect to database: %s", err)
			}

			log.Println("Connection to MongoDB established")
		}

		BeforeEach(func() {
			api = api2go.NewAPIWithBaseURL("v0", "http://localhost:31415")
			connectToMongoDB()
			err := db.DB("carshare").DropDatabase()
			Expect(err).ToNot(HaveOccurred())
			tripStorage := mongodb_storage.NewTripStorage(db)
			userStorage := mongodb_storage.NewUserStorage(db)
			carShareStorage := mongodb_storage.NewCarShareStorage(db)
			mockClock = clock.NewMock()
			api.AddResource(model.User{}, resource.UserResource{UserStorage: userStorage})
			api.AddResource(model.Trip{}, resource.TripResource{TripStorage: tripStorage, UserStorage: userStorage, CarShareStorage: carShareStorage, Clock: mockClock})
			api.AddResource(model.CarShare{}, resource.CarShareResource{CarShareStorage: carShareStorage, TripStorage: tripStorage, UserStorage: userStorage})
			rec = httptest.NewRecorder()
		})

		It("Creates a new user", func() {
			createUser()
		})
	})

	Describe("Using in memory data store", func() {
		BeforeEach(func() {
			api = api2go.NewAPIWithBaseURL("v0", "http://localhost:31415")
			tripStorage := in_memory_storage.NewTripStorage()
			userStorage := in_memory_storage.NewUserStorage()
			carShareStorage := in_memory_storage.NewCarShareStorage()
			mockClock = clock.NewMock()
			api.AddResource(model.User{}, resource.UserResource{UserStorage: userStorage})
			api.AddResource(model.Trip{}, resource.TripResource{TripStorage: tripStorage, UserStorage: userStorage, CarShareStorage: carShareStorage, Clock: mockClock})
			api.AddResource(model.CarShare{}, resource.CarShareResource{CarShareStorage: carShareStorage, TripStorage: tripStorage, UserStorage: userStorage})
			rec = httptest.NewRecorder()
		})

		It("Creates a new user", func() {
			createUser()
		})
	})

	/*var createCarShare = func() string {
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
		actual := rec.Body.String()
		id := extractIDFromResponse(actual)
		expected := strings.NewReplacer("<<id>>", id).Replace(`
		{
		  "data": {
		    "type": "carShares",
		    "id": "<<id>>",
		    "attributes": {
		      "name": "carShare1"
		    },
		    "relationships": {
		      "admins": {
		        "links": {
		          "self": "http://localhost:31415/v0/carShares/<<id>>/relationships/admins",
		          "related": "http://localhost:31415/v0/carShares/<<id>>/admins"
		        },
		        "data": []
		      },
		      "trips": {
		        "links": {
		          "self": "http://localhost:31415/v0/carShares/<<id>>/relationships/trips",
		          "related": "http://localhost:31415/v0/carShares/<<id>>/trips"
		        },
		        "data": []
		      }
		    }
		  }
		}
		`)
		Expect(actual).To(MatchJSON(expected))
		return id
	}

	It("Creates a new car share", func() {
		createCarShare()
	})

	var createTrip = func() string {
		rec = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/v0/trips", strings.NewReader(`
		{
		  "data": {
		    "type": "trips",
		    "attributes": {
		      "metres": 1000
		    }
		  }
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusCreated))
		actual := rec.Body.String()
		id := extractIDFromResponse(actual)
		expected := strings.NewReplacer("<<id>>", id).Replace(`
		{
		  "data": {
		    "type": "trips",
		    "id": "<<id>>",
		    "attributes": {
		      "metres": 1000,
		      "timestamp": "1970-01-01T00:00:00Z",
		      "scores": {}
		    },
		    "relationships": {
		      "carShare": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/<<id>>/relationships/carShare",
		          "related": "http://localhost:31415/v0/trips/<<id>>/carShare"
		        },
		        "data": null
		      },
		      "driver": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/<<id>>/relationships/driver",
		          "related": "http://localhost:31415/v0/trips/<<id>>/driver"
		        },
		        "data": null
		      },
		      "passengers": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/<<id>>/relationships/passengers",
		          "related": "http://localhost:31415/v0/trips/<<id>>/passengers"
		        },
		        "data": []
		      }
		    }
		  }
		}
		`)
		Expect(actual).To(MatchJSON(expected))
		return id
	}

	It("Creates a trip", func() {
		createTrip()
	})*/

	/*It("Adds a driver to a trip", func() {
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

		By("Loading the trip from the backend, it should have the user as the driver")

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
		      "metres": 1000,
		      "timestamp": "1970-01-01T00:00:00Z",
		      "scores": {}
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

	It("Links a trip to a car share", func() {
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
		      "metres": 1000,
		      "timestamp": "1970-01-01T00:00:00Z",
		      "scores": {}
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
		      }
		    }
		  }
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
		  "data": [
		    {
		      "type": "trips",
		      "id": "1"
		    }
		  ]
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
		        "metres": 1000,
		        "timestamp": "1970-01-01T00:00:00Z",
		        "scores": {}
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
		  "data": [
		    {
		      "type": "trips",
		      "id": "1"
		    }
		  ]
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
		        "metres": 1000,
		        "timestamp": "1970-01-01T00:00:00Z",
		        "scores": {}
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
		  "data": [
		    {
		      "type": "trips",
		      "id": "1"
		    }
		  ]
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

	It("Should be able to handle Scenario 1", func() {

		By("Creating a few users")

		createUser()

		rec = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/v0/users", strings.NewReader(`
		{
		  "data": {
		    "type": "users",
		    "attributes": {
		      "user-name": "paul"
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
		    "id": "2",
		    "attributes": {
		      "user-name": "paul"
		    }
		  }
		}
		`))

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/v0/users", strings.NewReader(`
		{
		  "data": {
		    "type": "users",
		    "attributes": {
		      "user-name": "john"
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
		    "id": "3",
		    "attributes": {
		      "user-name": "john"
		    }
		  }
		}
		`))

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/v0/users", strings.NewReader(`
		{
		  "data": {
		    "type": "users",
		    "attributes": {
		      "user-name": "angela"
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
		    "id": "4",
		    "attributes": {
		      "user-name": "angela"
		    }
		  }
		}
		`))

		By("Create a car share")

		createCarShare()

		By("Add a trip to the car share. Marvin drives with Paul and John as passengers")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/v0/trips", strings.NewReader(`
		{
		  "data": {
		    "type": "trips",
		    "attributes": {
		      "metres": 1
		    },
		    "relationships": {
		      "carShare": {
		        "data": {
		          "type": "carShares",
		          "id": "1"
		        }
		      },
		      "driver": {
		        "data": {
		          "type": "users",
		          "id": "1"
		        }
		      },
		      "passengers": {
		        "data": [
		          {
		            "type": "users",
		            "id": "2"
		          },
		          {
		            "type": "users",
		            "id": "3"
		          }
		        ]
		      }
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
		      "metres": 1,
		      "timestamp": "1970-01-01T00:00:00Z",
		      "scores": {
		        "1": {
		          "metres-as-driver": 1,
		          "metres-as-passenger": 0
		        },
		        "2": {
		          "metres-as-driver": 0,
		          "metres-as-passenger": 1
		        },
		        "3": {
		          "metres-as-driver": 0,
		          "metres-as-passenger": 1
		        }
		      }
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
		        "data": [
		          {
		            "type": "users",
		            "id": "2"
		          },
		          {
		            "type": "users",
		            "id": "3"
		          }
		        ]
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
		    },
		    {
		      "type": "users",
		      "id": "2",
		      "attributes": {
		        "user-name": "paul"
		      }
		    },
		    {
		      "type": "users",
		      "id": "3",
		      "attributes": {
		        "user-name": "john"
		      }
		    }
		  ]
		}
		`))

		By("Add another trip to the car share. Paul drives with Marvin and John as passengers")

		mockClock.Add(24 * time.Hour)

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/v0/trips", strings.NewReader(`
		{
		  "data": {
		    "type": "trips",
		    "attributes": {
		      "metres": 1
		    },
		    "relationships": {
		      "carShare": {
		        "data": {
		          "type": "carShares",
		          "id": "1"
		        }
		      },
		      "driver": {
		        "data": {
		          "type": "users",
		          "id": "2"
		        }
		      },
		      "passengers": {
		        "data": [
		          {
		            "type": "users",
		            "id": "1"
		          },
		          {
		            "type": "users",
		            "id": "3"
		          }
		        ]
		      }
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
		    "id": "2",
		    "attributes": {
		      "metres": 1,
		      "timestamp": "1970-01-02T00:00:00Z",
		      "scores": {
		        "1": {
		          "metres-as-driver": 1,
		          "metres-as-passenger": 1
		        },
		        "2": {
		          "metres-as-driver": 1,
		          "metres-as-passenger": 1
		        },
		        "3": {
		          "metres-as-driver": 0,
		          "metres-as-passenger": 2
		        }
		      }
		    },
		    "relationships": {
		      "carShare": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/2/relationships/carShare",
		          "related": "http://localhost:31415/v0/trips/2/carShare"
		        },
		        "data": {
		          "type": "carShares",
		          "id": "1"
		        }
		      },
		      "driver": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/2/relationships/driver",
		          "related": "http://localhost:31415/v0/trips/2/driver"
		        },
		        "data": {
		          "type": "users",
		          "id": "2"
		        }
		      },
		      "passengers": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/2/relationships/passengers",
		          "related": "http://localhost:31415/v0/trips/2/passengers"
		        },
		        "data": [
		          {
		            "type": "users",
		            "id": "1"
		          },
		          {
		            "type": "users",
		            "id": "3"
		          }
		        ]
		      }
		    }
		  },
		  "included": [
		    {
		      "type": "users",
		      "id": "2",
		      "attributes": {
		        "user-name": "paul"
		      }
		    },
		    {
		      "type": "users",
		      "id": "1",
		      "attributes": {
		        "user-name": "marvin"
		      }
		    },
		    {
		      "type": "users",
		      "id": "3",
		      "attributes": {
		        "user-name": "john"
		      }
		    }
		  ]
		}
		`))

		By("Add another trip to the car share. Paul drives with Marvin as the passenger. John isn't car sharing today")

		mockClock.Add(24 * time.Hour)

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("POST", "/v0/trips", strings.NewReader(`
		{
		  "data": {
		    "type": "trips",
		    "attributes": {
		      "metres": 1
		    },
		    "relationships": {
		      "carShare": {
		        "data": {
		          "type": "carShares",
		          "id": "1"
		        }
		      },
		      "driver": {
		        "data": {
		          "type": "users",
		          "id": "2"
		        }
		      },
		      "passengers": {
		        "data": [
		          {
		            "type": "users",
		            "id": "1"
		          }
		        ]
		      }
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
		    "id": "3",
		    "attributes": {
		      "metres": 1,
		      "timestamp": "1970-01-03T00:00:00Z",
		      "scores": {
		        "1": {
		          "metres-as-driver": 1,
		          "metres-as-passenger": 2
		        },
		        "2": {
		          "metres-as-driver": 2,
		          "metres-as-passenger": 1
		        },
		        "3": {
		          "metres-as-driver": 0,
		          "metres-as-passenger": 2
		        }
		      }
		    },
		    "relationships": {
		      "carShare": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/3/relationships/carShare",
		          "related": "http://localhost:31415/v0/trips/3/carShare"
		        },
		        "data": {
		          "type": "carShares",
		          "id": "1"
		        }
		      },
		      "driver": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/3/relationships/driver",
		          "related": "http://localhost:31415/v0/trips/3/driver"
		        },
		        "data": {
		          "type": "users",
		          "id": "2"
		        }
		      },
		      "passengers": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/3/relationships/passengers",
		          "related": "http://localhost:31415/v0/trips/3/passengers"
		        },
		        "data": [
		          {
		            "type": "users",
		            "id": "1"
		          }
		        ]
		      }
		    }
		  },
		  "included": [
		    {
		      "type": "users",
		      "id": "2",
		      "attributes": {
		        "user-name": "paul"
		      }
		    },
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
	})*/
})

/*
Extract the contents of the first id tag in a JSON string
*/
func extractIDFromResponse(response string) string {

	/*
		Split response string into two strings around the id tag

		{"data":{"type":"users","id":"582dbd234eb12720a3adb5d9","attribu...
		-> [{"data":{"type":"users","id":],
			 [582dbd234eb12720a3adb5d9","attribu...]
	*/
	tokens := strings.SplitAfterN(response, "id\":\"", 2)

	if len(tokens) != 2 {
		return "unable to find id in response"
	}

	/*
	 extract the id by stripping out everythinng from the first quote.

	 582dbd234eb12720a3adb5d9","attribu...
	 -> 582dbd234eb12720a3adb5d9
	*/
	id := tokens[1][0:strings.Index(tokens[1], "\"")]

	return id
}
