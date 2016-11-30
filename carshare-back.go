/*
RESTful API for the car share system
*/
package main

import (
	"fmt"
	"net/http"

	mgo "gopkg.in/mgo.v2"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resolver"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage/mongodb"
	"github.com/benbjohnson/clock"
	"github.com/julienschmidt/httprouter"
	"github.com/manyminds/api2go"
)

func main() {
	port := 31415
	api := api2go.NewAPIWithResolver("v0", &resolver.RequestURL{Port: port})
	db, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	api.UseMiddleware(
		func(c api2go.APIContexter, w http.ResponseWriter, r *http.Request) {
			c.Set("db", db)
		},
	)
	userStorage := &mongodb_storage.UserStorage{}
	tripStorage := &mongodb_storage.TripStorage{}
	carShareStorage := mongodb_storage.NewCarShareStorage(db)
	clock := clock.New()
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
			Clock:           clock,
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

	fmt.Printf("Listening on :%d", port)
	handler := api.Handler().(*httprouter.Router)

	http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}
