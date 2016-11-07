/*
RESTful API for the car share system
*/
package main

import (
	"fmt"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resolver"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/benbjohnson/clock"
	"github.com/julienschmidt/httprouter"
	"github.com/manyminds/api2go"
)

func main() {
	port := 31415
	api := api2go.NewAPIWithResolver("v0", &resolver.RequestURL{Port: port})
	tripStorage := storage.NewTripStorage()
	userStorage := storage.NewUserStorage()
	carShareStorage := storage.NewCarShareStorage()
	clock := clock.New()
	api.AddResource(model.User{}, resource.UserResource{UserStorage: userStorage})
	api.AddResource(model.Trip{}, resource.TripResource{TripStorage: tripStorage, UserStorage: userStorage, CarShareStorage: carShareStorage, Clock: clock})
	api.AddResource(model.CarShare{}, resource.CarShareResource{CarShareStorage: carShareStorage, TripStorage: tripStorage, UserStorage: userStorage})

	fmt.Printf("Listening on :%d", port)
	handler := api.Handler().(*httprouter.Router)

	http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}
