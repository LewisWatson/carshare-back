/*
RESTful API for the car share system
*/
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resolver"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage/mongodb"
	"github.com/benbjohnson/clock"
	"github.com/julienschmidt/httprouter"
	"github.com/manyminds/api2go"
	"gopkg.in/LewisWatson/firebase-jwt-auth.v1"

	"gopkg.in/alecthomas/kingpin.v2"
	mgo "gopkg.in/mgo.v2"
)

var (
	port   = kingpin.Flag("port", "Set port to bind to").Default("31415").Envar("CARSHARE_PORT").Int()
	mgoURL = kingpin.Flag("mgoURL", "URL to MongoDB server or seed server(s) for clusters").Default("localhost").Envar("CARSHARE_MGO_URL").URL()
	acao   = kingpin.Flag("cors", "Enable HTTP Access Control (CORS) for the specified URI").PlaceHolder("URI").Envar("CARSHARE_CORS_URI").String()
)

func main() {

	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("0.3.3").Author("Lewis Watson")
	kingpin.CommandLine.Help = "API for tracking car shares"
	kingpin.Parse()

	api := api2go.NewAPIWithResolver("v0", &resolver.RequestURL{Port: *port})

	log.Printf("connecting to mongodb server %s%s", (*mgoURL).Host, (*mgoURL).Path)
	db, err := mgo.Dial((*mgoURL).String())
	if err != nil {
		log.Fatalf("error connecting to mongodb server: %s", err)
	}

	if *acao != "" {
		log.Printf("enabling CORS access for %s", *acao)
		api.UseMiddleware(
			func(c api2go.APIContexter, w http.ResponseWriter, r *http.Request) {
				c.Set("db", db)
				w.Header().Set("Access-Control-Allow-Origin", *acao)
				w.Header().Set("Access-Control-Allow-Headers", "Authorization,content-type")
				w.Header().Set("Access-Control-Allow-Methods", "GET,PATCH,DELETE,OPTIONS")
			},
		)
	}

	userStorage := &mongodb.UserStorage{}
	carShareStorage := &mongodb.CarShareStorage{}
	tripStorage := &mongodb.TripStorage{}

	tokenVerifier, err := fireauth.New("ridesharelogger")
	if err != nil {
		log.Fatal(err)
	}

	api.AddResource(
		model.User{},
		resource.UserResource{
			UserStorage: userStorage,
		},
	)
	api.AddResource(
		model.Trip{},
		resource.TripResource{
			TripStorage:     tripStorage,
			UserStorage:     userStorage,
			CarShareStorage: carShareStorage,
			Clock:           clock.New(),
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

	log.Printf("listening on :%d", *port)
	err = http.ListenAndServe(
		fmt.Sprintf(":%d", *port),
		api.Handler().(*httprouter.Router),
	)

	if err != nil {
		log.Fatal(err)
	}
}
