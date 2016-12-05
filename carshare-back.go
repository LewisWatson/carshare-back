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

var mgoUrl = os.Getenv("CARSHARE_MGO_URL")
var portEnv = os.Getenv("CARSHARE_PORT")

func main() {

	port := getPort()
	api := api2go.NewAPIWithResolver("v0", &resolver.RequestURL{Port: port})

	db := connectToMgo()
	api.UseMiddleware(
		func(c api2go.APIContexter, w http.ResponseWriter, r *http.Request) {
			c.Set("db", db)
		},
	)

	userStorage := &mongodb_storage.UserStorage{}
	tripStorage := &mongodb_storage.TripStorage{}
	carShareStorage := &mongodb_storage.CarShareStorage{}

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
			UserStorage:     &mongodb_storage.UserStorage{},
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
	if mgoUrl == "" {
		mgoUrl = "localhost"
	}
	log.Printf("connecting to mongodb server via url: %s", mgoUrl)
	db, err := mgo.Dial(mgoUrl)
	if err != nil {
		panic(fmt.Sprintf("error connecting to mongodb server\n%s", err))
	}
	return db
}
