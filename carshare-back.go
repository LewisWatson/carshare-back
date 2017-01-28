/*
RESTful API for the car share system
*/
package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"

	mgo "gopkg.in/mgo.v2"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resolver"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage/mongodb"
	"github.com/benbjohnson/clock"
	"github.com/julienschmidt/httprouter"
	"github.com/manyminds/api2go"
)

var mgoURL = os.Getenv("CARSHARE_MGO_URL")
var portEnv = os.Getenv("CARSHARE_PORT")

func main() {

	port := getPort()
	api := api2go.NewAPIWithResolver("v0", &resolver.RequestURL{Port: port})

	db := connectToMgo()
	api.UseMiddleware(
		func(c api2go.APIContexter, w http.ResponseWriter, r *http.Request) {
			c.Set("db", db)
			w.Header().Set("Access-Control-Allow-Origin", "*")
			w.Header().Set("Access-Control-Allow-Headers", "content-type")
			w.Header().Set("Access-Control-Allow-Methods", "GET,PATCH,DELETE,OPTIONS")
		},
	)

	userStorage := &mongodb.UserStorage{}
	carShareStorage := &mongodb.CarShareStorage{}
	tripStorage := &mongodb.TripStorage{}

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
		},
	)

	log.Printf("listening on :%d", port)
	err := http.ListenAndServe(
		fmt.Sprintf(":%d", port),
		api.Handler().(*httprouter.Router),
	)

	if err != nil {
		log.Fatal(err)
	}
}

func getPort() int {
	var err error
	port := 31415
	if portEnv != "" {
		port, err = strconv.Atoi(portEnv)
		if err != nil {
			panic(fmt.Sprintf("unable to parse port environmental variable\n%s", err))
		}
	}
	return port
}

func connectToMgo() *mgo.Session {
	if mgoURL == "" {
		mgoURL = "localhost"
	}
	log.Printf("connecting to mongodb server viaaaa url: %s", mgoURL)
	db, err := mgo.Dial(mgoURL)
	if err != nil {
		panic(fmt.Sprintf("error connecting to mongodb server\n%s", err))
	}
	return db
}
