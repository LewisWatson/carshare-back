package main_test

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"

	mgo "gopkg.in/mgo.v2"
	dockertest "gopkg.in/ory-am/dockertest.v3"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage/mongodb"
	"github.com/benbjohnson/clock"
	"github.com/manyminds/api2go"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"gopkg.in/jose.v1/jwt"
)

var (
	db                *mgo.Session
	pool              *dockertest.Pool
	containerResource *dockertest.Resource
)

type mockTokenVerifier struct {
	Claims jwt.Claims
	Error  error
}

func (mtv mockTokenVerifier) Verify(accessToken string) (userID string, claims jwt.Claims, err error) {
	return mtv.Claims.Get("sub").(string), mtv.Claims, mtv.Error
}

var _ = AfterSuite(func() {

	fmt.Println()

	if db != nil {
		log.Println("Closing connection to MongoDB")
		db.Close()
	}

	if pool != nil {
		log.Println("Purging containers")
		if err := pool.Purge(containerResource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}
	}
})

// there are a lot of functions because each test can be run individually and sets up the complete
// environment. That is because we run all the specs randomized.
var _ = Describe("The CarShareBack API", func() {

	var (
		rec           *httptest.ResponseRecorder
		mockClock     *clock.Mock
		tokenVerifier mockTokenVerifier
	)

	var createUser = func(name string) string {
		rec = httptest.NewRecorder()
		req, err := http.NewRequest("POST", "/v0/users", strings.NewReader(`
		{
			"data": {
				"type": "users",
				"attributes": {
					"user-name": "`+name+`"
				}
			}
		}
		`))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		Expect(rec.Code).To(Equal(http.StatusCreated))
		actual := rec.Body.String()
		id := extractIDFromResponse(actual)
		Expect(actual).To(MatchJSON(`
		{
			"data": {
				"type": "users",
				"id": "` + id + `",
				"attributes": {
					"user-name": "` + name + `"
				}
			}
		}
		`))
		return id
	}

	var createCarShare = func(userID string) string {

		tokenVerifier.Claims.Set("sub", userID)

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
		expected := strings.NewReplacer("<<id>>", id, "<<userID>>", userID).Replace(`
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
		        "data": [
							{
								"type": "users",
								"id": "<<userID>>"
							}
						]
		      },
					"members": {
		        "links": {
		          "self": "http://localhost:31415/v0/carShares/<<id>>/relationships/members",
		          "related": "http://localhost:31415/v0/carShares/<<id>>/members"
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
		  },
			"included": [
				{
					"type": "users",
					"id": "<<userID>>",
					"attributes": {
						"user-name": "example"
					}
				}
			]
		}
		`)
		Expect(actual).To(MatchJSON(expected))
		return id
	}

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

	var addDriverToTrip = func() {
		userID := createUser("marvin")
		tripID := createTrip()

		By("Adding a driver to a trip with PATCH")

		replacer := strings.NewReplacer("<<trip-id>>", tripID, "<<user-id>>", userID)
		requestUrl := replacer.Replace("/v0/trips/<<trip-id>>")
		requestBody := replacer.Replace(`
		{
			"data": {
				"type": "trips",
				"id": "<<trip-id>>",
				"attributes": {},
				"relationships": {
					"driver": {
						"data": {
							"type": "users",
							"id": "<<user-id>>"
						}
					}
				}
			}
		}
		`)

		req, err := http.NewRequest("PATCH", requestUrl, strings.NewReader(requestBody))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		// Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the trip from the backend, it should have the user as the driver")

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", "/v0/trips/"+tripID, nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(replacer.Replace(`
		{
			"data": {
				"type": "trips",
				"id": "<<trip-id>>",
				"attributes": {
					"metres": 1000,
					"timestamp": "1970-01-01T00:00:00Z",
					 "scores": {
							"<<user-id>>": {
								"metres-as-driver": 1000,
								"metres-as-passenger": 0
							}
						}
				},
				"relationships": {
					"carShare": {
						"links": {
							"self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/carShare",
							"related": "http://localhost:31415/v0/trips/<<trip-id>>/carShare"
						},
						"data": null
					},
					"driver": {
						"links": {
							"self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/driver",
							"related": "http://localhost:31415/v0/trips/<<trip-id>>/driver"
						},
						"data": {
							"type": "users",
							"id": "<<user-id>>"
						}
					},
					"passengers": {
						"links": {
							"self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/passengers",
							"related": "http://localhost:31415/v0/trips/<<trip-id>>/passengers"
						},
						"data": []
					}
				}
			},
			"included": [
				{
					"type": "users",
					"id": "<<user-id>>",
					"attributes": {
						"user-name": "marvin"
					}
				}
			]
		}
		`)))
	}

	var linkTripToCarShare = func() {
		userID := createUser("example")
		carShareID := createCarShare(userID)
		tripID := createTrip()

		By("Adding a carShare to a trip with PATCH")

		replacer := strings.NewReplacer("<<trip-id>>", tripID, "<<carshare-id>>", carShareID)
		requestUrl := replacer.Replace("/v0/trips/<<trip-id>>")
		requestBody := replacer.Replace(`
		{
		  "data": {
		    "type": "trips",
		    "id": "<<trip-id>>",
		    "attributes": {},
		    "relationships": {
		      "carShare": {
		        "data": {
		          "type": "carShares",
		          "id": "<<carshare-id>>"
		        }
		      }
		    }
		  }
		}
		`)

		req, err := http.NewRequest("PATCH", requestUrl, strings.NewReader(requestBody))
		Expect(err).ToNot(HaveOccurred())
		api.Handler().ServeHTTP(rec, req)
		// Expect(rec.Code).To(Equal(http.StatusNoContent))

		By("Loading the trip, it should have the car share")

		getUrl := replacer.Replace("/v0/trips/<<trip-id>>")
		expected := replacer.Replace(`
		{
		  "data": {
		    "type": "trips",
		    "id": "<<trip-id>>",
		    "attributes": {
		      "metres": 1000,
		      "timestamp": "1970-01-01T00:00:00Z",
		      "scores": {}
		    },
		    "relationships": {
		      "carShare": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/carShare",
		          "related": "http://localhost:31415/v0/trips/<<trip-id>>/carShare"
		        },
		        "data": {
		          "type": "carShares",
		          "id": "<<carshare-id>>"
		        }
		      },
		      "driver": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/driver",
		          "related": "http://localhost:31415/v0/trips/<<trip-id>>/driver"
		        },
		        "data": null
		      },
		      "passengers": {
		        "links": {
		          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/passengers",
		          "related": "http://localhost:31415/v0/trips/<<trip-id>>/passengers"
		        },
		        "data": []
		      }
		    }
		  }
		}
		`)

		rec = httptest.NewRecorder()
		req, err = http.NewRequest("GET", getUrl, nil)
		api.Handler().ServeHTTP(rec, req)
		Expect(err).ToNot(HaveOccurred())
		Expect(rec.Body.String()).To(MatchJSON(expected))
	}

	// var addTripToCarShare = func() {
	// 	userID := createUser("example")
	// 	carShareID := createCarShare(userID)
	// 	tripID := createTrip()

	// 	replacer := strings.NewReplacer(
	// 		"<<trip-id>>", tripID,
	// 		"<<carshare-id>>", carShareID,
	// 	)

	// 	By("Adding a trip with POST")

	// 	requestUrl := replacer.Replace("/v0/carShares/<<carshare-id>>/relationships/trips")
	// 	requestBody := strings.NewReader(replacer.Replace(`
	// 	{
	// 	  "data": [
	// 	    {
	// 	      "type": "trips",
	// 	      "id": "<<trip-id>>"
	// 	    }
	// 	  ]
	// 	}
	// 	`))
	// 	req, err := http.NewRequest("POST", requestUrl, requestBody)
	// 	Expect(err).ToNot(HaveOccurred())
	// 	api.Handler().ServeHTTP(rec, req)
	// 	Expect(rec.Code).To(Equal(http.StatusCreated))

	// 	By("Loading the car share from the backend, it should have the trip")

	// 	rec = httptest.NewRecorder()
	// 	req, err = http.NewRequest("GET", replacer.Replace("/v0/carShares/<<carshare-id>>"), nil)
	// 	api.Handler().ServeHTTP(rec, req)
	// 	Expect(err).ToNot(HaveOccurred())
	// 	Expect(rec.Body.String()).To(MatchJSON(replacer.Replace(`
	// 	{
	// 	  "data": {
	// 	    "type": "carShares",
	// 	    "id": "<<carshare-id>>",
	// 	    "attributes": {
	// 	      "name": "carShare1"
	// 	    },
	// 	    "relationships": {
	// 	      "admins": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/admins",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/admins"
	// 	        },
	// 	        "data": []
	// 	      },
	// 				"members": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/members",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/members"
	// 	        },
	// 	        "data": []
	// 	      },
	// 	      "trips": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/trips",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/trips"
	// 	        },
	// 	        "data": [
	// 	          {
	// 	            "type": "trips",
	// 	            "id": "<<trip-id>>"
	// 	          }
	// 	        ]
	// 	      }
	// 	    }
	// 	  },
	// 	  "included": [
	// 	    {
	// 	      "type": "trips",
	// 	      "id": "<<trip-id>>",
	// 	      "attributes": {
	// 	        "metres": 1000,
	// 	        "timestamp": "1970-01-01T00:00:00Z",
	// 	        "scores": {}
	// 	      },
	// 	      "relationships": {
	// 	         "carShare": {
	//                 "links": {
	//                   "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/carShare",
	//                   "related": "http://localhost:31415/v0/trips/<<trip-id>>/carShare"
	//                 },
	//                 "data": {
	//                   "type": "carShares",
	//                   "id": "<<carshare-id>>"
	//                 }
	//               },
	// 	        "driver": {
	// 	          "links": {
	// 	            "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/driver",
	// 	            "related": "http://localhost:31415/v0/trips/<<trip-id>>/driver"
	// 	          },
	// 	          "data": null
	// 	        },
	// 	        "passengers": {
	// 	          "links": {
	// 	            "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/passengers",
	// 	            "related": "http://localhost:31415/v0/trips/<<trip-id>>/passengers"
	// 	          },
	// 	          "data": []
	// 	        }
	// 	      }
	// 	    }
	// 	  ]
	// 	}
	// 	`)))
	// }

	// var replaceTrips = func(carShareID string, tripID string) {
	// 	By("Replacing trip relationship with PATCH")

	// 	replacer := strings.NewReplacer("<<carshare-id>>", carShareID, "<<trip-id>>", tripID)

	// 	requestUrl := replacer.Replace("/v0/carShares/<<carshare-id>>/relationships/trips")
	// 	requestBody := strings.NewReader(replacer.Replace(`
	// 	{
	// 	  "data": [
	// 	    {
	// 	      "type": "trips",
	// 	      "id": "<<trip-id>>"
	// 	    }
	// 	  ]
	// 	}
	// 	`))
	// 	rec = httptest.NewRecorder()
	// 	req, err := http.NewRequest("PATCH", requestUrl, requestBody)
	// 	Expect(err).ToNot(HaveOccurred())
	// 	api.Handler().ServeHTTP(rec, req)
	// 	Expect(rec.Code).To(Equal(http.StatusNoContent))

	// 	By("Loading the car share from the backend, it should have the relationship")

	// 	rec = httptest.NewRecorder()
	// 	req, err = http.NewRequest("GET", replacer.Replace("/v0/carShares/<<carshare-id>>"), nil)
	// 	api.Handler().ServeHTTP(rec, req)
	// 	Expect(err).ToNot(HaveOccurred())
	// 	Expect(rec.Body.String()).To(MatchJSON(replacer.Replace(`
	// 	{
	// 	  "data": {
	// 	    "type": "carShares",
	// 	    "id": "<<carshare-id>>",
	// 	    "attributes": {
	// 	      "name": "carShare1"
	// 	    },
	// 	    "relationships": {
	// 	      "admins": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/admins",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/admins"
	// 	        },
	// 	        "data": []
	// 	      },
	// 				"members": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/members",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/members"
	// 	        },
	// 	        "data": []
	// 	      },
	// 	      "trips": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/trips",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/trips"
	// 	        },
	// 	        "data": [
	// 	          {
	// 	            "type": "trips",
	// 	            "id": "<<trip-id>>"
	// 	          }
	// 	        ]
	// 	      }
	// 	    }
	// 	  },
	// 	  "included": [
	// 	    {
	// 	      "type": "trips",
	// 	      "id": "<<trip-id>>",
	// 	      "attributes": {
	// 	        "metres": 1000,
	// 	        "timestamp": "1970-01-01T00:00:00Z",
	// 	        "scores": {}
	// 	      },
	// 	      "relationships": {
	// 	        "carShare": {
	// 	          "links": {
	// 	            "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/carShare",
	// 	            "related": "http://localhost:31415/v0/trips/<<trip-id>>/carShare"
	// 	          },
	// 	          "data": {
	//                   "type": "carShares",
	//                   "id": "<<carshare-id>>"
	//                 }
	// 	        },
	// 	        "driver": {
	// 	          "links": {
	// 	            "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/driver",
	// 	            "related": "http://localhost:31415/v0/trips/<<trip-id>>/driver"
	// 	          },
	// 	          "data": null
	// 	        },
	// 	        "passengers": {
	// 	          "links": {
	// 	            "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/passengers",
	// 	            "related": "http://localhost:31415/v0/trips/<<trip-id>>/passengers"
	// 	          },
	// 	          "data": []
	// 	        }
	// 	      }
	// 	    }
	// 	  ]
	// 	}
	// 	`)))
	// }

	// var deleteCarShareTrip = func() {

	// 	userID := createUser("example")

	// 	By("create a car share")
	// 	carShareID := createCarShare(userID)

	// 	By("create a trip")
	// 	tripID := createTrip()

	// 	replacer := strings.NewReplacer("<<carshare-id>>", carShareID, "<<trip-id>>", tripID)

	// 	By(replacer.Replace("add trip <<trip-id>> to car share <<carshare-id>>"))
	// 	replaceTrips(carShareID, tripID)

	// 	By(replacer.Replace("delete trip <<trip-id>> from car share <<carshare-id>>"))

	// 	rec = httptest.NewRecorder()
	// 	requestUrl := replacer.Replace("/v0/carShares/<<carshare-id>>/relationships/trips")
	// 	requestBody := strings.NewReader(replacer.Replace(`
	// 	{
	// 	  "data": [
	// 	    {
	// 	      "type": "trips",
	// 	      "id": "<<trip-id>>"
	// 	    }
	// 	  ]
	// 	}
	// 	`))
	// 	req, err := http.NewRequest("DELETE", requestUrl, requestBody)
	// 	Expect(err).ToNot(HaveOccurred())
	// 	api.Handler().ServeHTTP(rec, req)
	// 	Expect(rec.Code).To(Equal(http.StatusNoContent))

	// 	By(replacer.Replace("check that trip <<trip-id>> is no longer in car share <<carshare-id>>"))

	// 	rec = httptest.NewRecorder()
	// 	req, err = http.NewRequest("GET", replacer.Replace("/v0/carShares/<<carshare-id>>"), nil)
	// 	api.Handler().ServeHTTP(rec, req)
	// 	Expect(err).ToNot(HaveOccurred())
	// 	Expect(rec.Body.String()).To(MatchJSON(replacer.Replace(`
	// 	{
	// 	  "data": {
	// 	    "type": "carShares",
	// 	    "id": "<<carshare-id>>",
	// 	    "attributes": {
	// 	      "name": "carShare1"
	// 	    },
	// 	    "relationships": {
	// 	      "admins": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/admins",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/admins"
	// 	        },
	// 	        "data": []
	// 	      },
	// 				"members": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/members",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/members"
	// 	        },
	// 	        "data": []
	// 	      },
	// 	      "trips": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/carShares/<<carshare-id>>/relationships/trips",
	// 	          "related": "http://localhost:31415/v0/carShares/<<carshare-id>>/trips"
	// 	        },
	// 	        "data": []
	// 	      }
	// 	    }
	// 	  }
	// 	}
	// 	`)))
	// }

	// var scenarioOne = func() {

	// 	By("Creating a few users")
	// 	marvinID := createUser("marvin")
	// 	paulID := createUser("paul")
	// 	johnID := createUser("john")
	// 	angelaID := createUser("angela")

	// 	By("Create a car share")
	// 	carShareID := createCarShare(marvinID)

	// 	replacer := strings.NewReplacer(
	// 		"<<carshare-id>>", carShareID,
	// 		"<<marvin-id>>", marvinID,
	// 		"<<paul-id>>", paulID,
	// 		"<<john-id>>", johnID,
	// 		"<<angela-id>>", angelaID,
	// 	)

	// 	By("Add a trip to the car share. Marvin drives with Paul and John as passengers")
	// 	rec = httptest.NewRecorder()
	// 	req, err := http.NewRequest(
	// 		"POST",
	// 		"/v0/trips",
	// 		strings.NewReader(replacer.Replace(`
	// 			{
	// 			  "data": {
	// 			    "type": "trips",
	// 			    "attributes": {
	// 			      "metres": 1
	// 			    },
	// 			    "relationships": {
	// 			      "carShare": {
	// 			        "data": {
	// 			          "type": "carShares",
	// 			          "id": "<<carshare-id>>"
	// 			        }
	// 			      },
	// 			      "driver": {
	// 			        "data": {
	// 			          "type": "users",
	// 			          "id": "<<marvin-id>>"
	// 			        }
	// 			      },
	// 			      "passengers": {
	// 			        "data": [
	// 			          {
	// 			            "type": "users",
	// 			            "id": "<<paul-id>>"
	// 			          },
	// 			          {
	// 			            "type": "users",
	// 			            "id": "<<john-id>>"
	// 			          }
	// 			        ]
	// 			      }
	// 			    }
	// 			  }
	// 			}
	// 			`)))
	// 	Expect(err).ToNot(HaveOccurred())
	// 	api.Handler().ServeHTTP(rec, req)
	// 	// Expect(rec.Code).To(Equal(http.StatusCreated))
	// 	actual := rec.Body.String()
	// 	tripID := extractIDFromResponse(actual)
	// 	expected := strings.NewReplacer("<<trip-id>>", tripID).Replace(replacer.Replace(`
	// 	{
	// 	  "data": {
	// 	    "type": "trips",
	// 	    "id": "<<trip-id>>",
	// 	    "attributes": {
	// 	      "metres": 1,
	// 	      "timestamp": "1970-01-01T00:00:00Z",
	// 	      "scores": {
	// 	        "<<marvin-id>>": {
	// 	          "metres-as-driver": 1,
	// 	          "metres-as-passenger": 0
	// 	        },
	// 	        "<<paul-id>>": {
	// 	          "metres-as-driver": 0,
	// 	          "metres-as-passenger": 1
	// 	        },
	// 	        "<<john-id>>": {
	// 	          "metres-as-driver": 0,
	// 	          "metres-as-passenger": 1
	// 	        }
	// 	      }
	// 	    },
	// 	    "relationships": {
	// 	      "carShare": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/carShare",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/carShare"
	// 	        },
	// 	        "data": {
	// 	          "type": "carShares",
	// 	          "id": "<<carshare-id>>"
	// 	        }
	// 	      },
	// 	      "driver": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/driver",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/driver"
	// 	        },
	// 	        "data": {
	// 	          "type": "users",
	// 	          "id": "<<marvin-id>>"
	// 	        }
	// 	      },
	// 	      "passengers": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/passengers",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/passengers"
	// 	        },
	// 	        "data": [
	// 	          {
	// 	            "type": "users",
	// 	            "id": "<<paul-id>>"
	// 	          },
	// 	          {
	// 	            "type": "users",
	// 	            "id": "<<john-id>>"
	// 	          }
	// 	        ]
	// 	      }
	// 	    }
	// 	  },
	// 	  "included": [
	// 	    {
	// 	      "type": "users",
	// 	      "id": "<<marvin-id>>",
	// 	      "attributes": {
	// 	        "user-name": "marvin"
	// 	      }
	// 	    },
	// 	    {
	// 	      "type": "users",
	// 	      "id": "<<paul-id>>",
	// 	      "attributes": {
	// 	        "user-name": "paul"
	// 	      }
	// 	    },
	// 	    {
	// 	      "type": "users",
	// 	      "id": "<<john-id>>",
	// 	      "attributes": {
	// 	        "user-name": "john"
	// 	      }
	// 	    }
	// 	  ]
	// 	}
	// 	`))
	// 	Expect(actual).To(MatchJSON(expected))

	// 	By("Add another trip to the car share. Paul drives with Marvin and John as passengers")
	// 	mockClock.Add(24 * time.Hour)
	// 	rec = httptest.NewRecorder()
	// 	req, err = http.NewRequest(
	// 		"POST",
	// 		"/v0/trips",
	// 		strings.NewReader(replacer.Replace(`
	// 			{
	// 			  "data": {
	// 			    "type": "trips",
	// 			    "attributes": {
	// 			      "metres": 1
	// 			    },
	// 			    "relationships": {
	// 			      "carShare": {
	// 			        "data": {
	// 			          "type": "carShares",
	// 			          "id": "<<carshare-id>>"
	// 			        }
	// 			      },
	// 			      "driver": {
	// 			        "data": {
	// 			          "type": "users",
	// 			          "id": "<<paul-id>>"
	// 			        }
	// 			      },
	// 			      "passengers": {
	// 			        "data": [
	// 			          {
	// 			            "type": "users",
	// 			            "id": "<<marvin-id>>"
	// 			          },
	// 			          {
	// 			            "type": "users",
	// 			            "id": "<<john-id>>"
	// 			          }
	// 			        ]
	// 			      }
	// 			    }
	// 			  }
	// 			}
	// 			`)))
	// 	Expect(err).ToNot(HaveOccurred())
	// 	api.Handler().ServeHTTP(rec, req)
	// 	// Expect(rec.Code).To(Equal(http.StatusCreated))
	// 	actual = rec.Body.String()
	// 	tripID = extractIDFromResponse(actual)
	// 	expected = strings.NewReplacer("<<trip-id>>", extractIDFromResponse(actual)).Replace(replacer.Replace(`
	// 	{
	// 	  "data": {
	// 	    "type": "trips",
	// 	    "id": "<<trip-id>>",
	// 			"attributes": {
	// 	      "metres": 1,
	// 	      "timestamp": "1970-01-02T00:00:00Z",
	// 	      "scores": {
	// 	        "<<marvin-id>>": {
	// 	          "metres-as-driver": 1,
	// 	          "metres-as-passenger": 1
	// 	        },
	// 	        "<<paul-id>>": {
	// 	          "metres-as-driver": 1,
	// 	          "metres-as-passenger": 1
	// 	        },
	// 	        "<<john-id>>": {
	// 	          "metres-as-driver": 0,
	// 	          "metres-as-passenger": 2
	// 	        }
	// 	      }
	// 	    },
	// 	    "relationships": {
	// 	      "carShare": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/carShare",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/carShare"
	// 	        },
	// 	        "data": {
	// 	          "type": "carShares",
	// 	          "id": "<<carshare-id>>"
	// 	        }
	// 	      },
	// 	      "driver": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/driver",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/driver"
	// 	        },
	// 	        "data": {
	// 	          "type": "users",
	// 	          "id": "<<paul-id>>"
	// 	        }
	// 	      },
	// 	      "passengers": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/passengers",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/passengers"
	// 	        },
	// 	        "data": [
	// 	          {
	// 	            "type": "users",
	// 	            "id": "<<marvin-id>>"
	// 	          },
	// 	          {
	// 	            "type": "users",
	// 	            "id": "<<john-id>>"
	// 	          }
	// 	        ]
	// 	      }
	// 	    }
	// 	  },
	// 	  "included": [
	// 			{
	// 				"type": "users",
	// 				"id": "<<paul-id>>",
	// 				"attributes": {
	// 					"user-name": "paul"
	// 				}
	// 			},
	// 			{
	// 	      "type": "users",
	// 	      "id": "<<marvin-id>>",
	// 	      "attributes": {
	// 	        "user-name": "marvin"
	// 	      }
	// 	    },
	// 	    {
	// 	      "type": "users",
	// 	      "id": "<<john-id>>",
	// 	      "attributes": {
	// 	        "user-name": "john"
	// 	      }
	// 	    }
	// 	  ]
	// 	}
	// 	`))
	// 	Expect(actual).To(MatchJSON(expected))

	// 	By("Add another trip to the car share. Paul drives with Marvin as the passenger. John isn't car sharing today")
	// 	mockClock.Add(24 * time.Hour)
	// 	rec = httptest.NewRecorder()
	// 	req, err = http.NewRequest(
	// 		"POST",
	// 		"/v0/trips",
	// 		strings.NewReader(replacer.Replace(`
	// 		{
	// 		  "data": {
	// 		    "type": "trips",
	// 		    "attributes": {
	// 		      "metres": 1
	// 		    },
	// 		    "relationships": {
	// 		      "carShare": {
	// 		        "data": {
	// 		          "type": "carShares",
	// 		          "id": "<<carshare-id>>"
	// 		        }
	// 		      },
	// 		      "driver": {
	// 		        "data": {
	// 		          "type": "users",
	// 		          "id": "<<paul-id>>"
	// 		        }
	// 		      },
	// 		      "passengers": {
	// 		        "data": [
	// 		          {
	// 		            "type": "users",
	// 		            "id": "<<marvin-id>>"
	// 		          }
	// 		        ]
	// 		      }
	// 		    }
	// 		  }
	// 		}
	// 		`)))
	// 	Expect(err).ToNot(HaveOccurred())
	// 	api.Handler().ServeHTTP(rec, req)
	// 	// Expect(rec.Code).To(Equal(http.StatusCreated))
	// 	actual = rec.Body.String()
	// 	tripID = extractIDFromResponse(actual)
	// 	expected = strings.NewReplacer("<<trip-id>>", tripID).Replace(replacer.Replace(`
	// 	{
	// 	  "data": {
	// 	    "type": "trips",
	// 	    "id": "<<trip-id>>",
	// 	    "attributes": {
	// 	      "metres": 1,
	// 	      "timestamp": "1970-01-03T00:00:00Z",
	// 	      "scores": {
	// 	        "<<marvin-id>>": {
	// 	          "metres-as-driver": 1,
	// 	          "metres-as-passenger": 2
	// 	        },
	// 	        "<<paul-id>>": {
	// 	          "metres-as-driver": 2,
	// 	          "metres-as-passenger": 1
	// 	        },
	// 	        "<<john-id>>": {
	// 	          "metres-as-driver": 0,
	// 	          "metres-as-passenger": 2
	// 	        }
	// 	      }
	// 	    },
	// 	    "relationships": {
	// 	      "carShare": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/carShare",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/carShare"
	// 	        },
	// 	        "data": {
	// 	          "type": "carShares",
	// 	          "id": "<<carshare-id>>"
	// 	        }
	// 	      },
	// 	      "driver": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/driver",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/driver"
	// 	        },
	// 	        "data": {
	// 	          "type": "users",
	// 	          "id": "<<paul-id>>"
	// 	        }
	// 	      },
	// 	      "passengers": {
	// 	        "links": {
	// 	          "self": "http://localhost:31415/v0/trips/<<trip-id>>/relationships/passengers",
	// 	          "related": "http://localhost:31415/v0/trips/<<trip-id>>/passengers"
	// 	        },
	// 	        "data": [
	// 	          {
	// 	            "type": "users",
	// 	            "id": "<<marvin-id>>"
	// 	          }
	// 	        ]
	// 	      }
	// 	    }
	// 	  },
	// 	  "included": [
	// 	    {
	// 	      "type": "users",
	// 	      "id": "<<paul-id>>",
	// 	      "attributes": {
	// 	        "user-name": "paul"
	// 	      }
	// 	    },
	// 	    {
	// 	      "type": "users",
	// 	      "id": "<<marvin-id>>",
	// 	      "attributes": {
	// 	        "user-name": "marvin"
	// 	      }
	// 	    }
	// 	  ]
	// 	}
	// 	`))
	// 	Expect(actual).To(MatchJSON(expected))
	// }

	// Describe("Using in memory data store", func() {
	// 	BeforeEach(func() {
	// 		api = api2go.NewAPIWithBaseURL("v0", "http://localhost:31415")
	// 		userStorage := memory.NewUserStorage()
	// 		carShareStorage := memory.NewCarShareStorage()
	// 		tripStorage := memory.NewTripStorage()
	// 		mockClock = clock.NewMock()
	// 		mockTokenVerifier := mockTokenVerifier{}
	// 		mockTokenVerifier.Claims = make(jwt.Claims)
	// 		mockTokenVerifier.Claims.Set("sub", "blah")
	// 		api.AddResource(model.User{},
	// 			resource.UserResource{UserStorage: userStorage})
	// 		api.AddResource(model.Trip{},
	// 			resource.TripResource{
	// 				TripStorage:     tripStorage,
	// 				UserStorage:     userStorage,
	// 				CarShareStorage: carShareStorage,
	// 				Clock:           mockClock,
	// 			})
	// 		api.AddResource(model.CarShare{},
	// 			resource.CarShareResource{
	// 				CarShareStorage: carShareStorage,
	// 				TripStorage:     tripStorage,
	// 				UserStorage:     userStorage,
	// 				TokenVerifier:   mockTokenVerifier,
	// 			})
	// 		rec = httptest.NewRecorder()
	// 	})

	// 	It("Crea*tes a new user", func() {
	// 		createUser("marvin")
	// 	})

	// 	It("Creates a new car share", func() {
	// 		createCarShare()
	// 	})

	// 	It("Creates a trip", func() {
	// 		createTrip()
	// 	})

	// 	It("Adds a driver to a trip", func() {
	// 		addDriverToTrip()
	// 	})

	// 	It("Links a trip to a car share", func() {
	// 		linkTripToCarShare()
	// 	})

	// 	It("Adds a trip to a car share", func() {
	// 		addTripToCarShare()
	// 	})

	// 	It("Replaces car share's trips", func() {
	// 		carShareID := createCarShare()
	// 		tripID := createTrip()
	// 		replaceTrips(carShareID, tripID)
	// 	})

	// 	It("Deletes a car share trip", func() {
	// 		deleteCarShareTrip()
	// 	})

	// 	It("Should be able to handle Scenario 1", func() {
	// 		scenarioOne()
	// 	})
	// })

	Describe("Using MongoDB data store", func() {

		BeforeEach(func() {
			api = api2go.NewAPIWithBaseURL("v0", "http://localhost:31415")
			db, pool, containerResource = mongodb.ConnectToMongoDB(db, pool, containerResource)
			err := db.DB("carshare").DropDatabase()
			Expect(err).ToNot(HaveOccurred())
			userStorage := &mongodb.UserStorage{}
			tripStorage := &mongodb.TripStorage{}
			carShareStorage := &mongodb.CarShareStorage{}
			mockClock = clock.NewMock()
			tokenVerifier = mockTokenVerifier{}
			tokenVerifier.Claims = make(jwt.Claims)
			tokenVerifier.Claims.Set("sub", "example user ID")
			api.AddResource(
				model.User{},
				resource.UserResource{
					UserStorage: userStorage,
				},
			)
			api.AddResource(model.Trip{},
				resource.TripResource{
					TripStorage:     tripStorage,
					UserStorage:     userStorage,
					CarShareStorage: carShareStorage,
					Clock:           mockClock,
				},
			)
			api.AddResource(
				model.CarShare{},
				resource.CarShareResource{
					CarShareStorage: carShareStorage,
					TripStorage:     tripStorage,
					UserStorage:     userStorage,
					TokenVerifier:   tokenVerifier,
				},
			)
			api.UseMiddleware(
				func(c api2go.APIContexter, w http.ResponseWriter, r *http.Request) {
					c.Set("db", db)
				},
			)
			rec = httptest.NewRecorder()
		})

		It("Creates a new user", func() {
			createUser("marvin")
		})

		// It("Creates a new car share", func() {
		// 	userID := createUser("example user")
		// 	createCarShare(userID)
		// })

		It("Creates a trip", func() {
			createTrip()
		})

		It("Adds a driver to a trip", func() {
			addDriverToTrip()
		})

		It("Links a trip to a car share", func() {
			linkTripToCarShare()
		})

		// It("Adds a trip to a car share", func() {
		// 	addTripToCarShare()
		// })

		// It("Replaces car share's trips", func() {
		// 	userID := createUser("example user")
		// 	carShareID := createCarShare(userID)
		// 	tripID := createTrip()
		// 	replaceTrips(carShareID, tripID)
		// })

		// It("Deletes a car share trip", func() {
		// 	deleteCarShareTrip()
		// })

		// It("Should be able to handle Scenario 1", func() {
		// 	scenarioOne()
		// })
	})
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
