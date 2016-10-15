/*
RESTful API for the car share system

Create a new trip:
	curl -X POST http://localhost:31415/v0/trips -d '{"data" : {"type" : "trips" , "attributes": {"metres-as-driver" : 0, "metres-as-passenger" : 0 }}}'

List tips:
	curl -X GET http://localhost:31415/v0/trips

List paginated trips:
	curl -X GET 'http://localhost:31415/v0/trips?page\[offset\]=0&page\[limit\]=2'
OR
	curl -X GET 'http://localhost:31415/v0/trips?page\[number\]=1&page\[size\]=2'

Update:
	curl -vX PATCH http://localhost:31415/v0/trips/1 -d '{ "data" : {"type" : "trips", "id": "1", "attributes": {"metres-as-driver" : 1, "metres-as-passenger" : 2}}}'

Delete:
	curl -vX DELETE http://localhost:31415/v0/trips/2

Create a new carShare:
	curl -X POST http://localhost:31415/v0/carShares -d '{"data" : {"type" : "carShares" , "attributes": {"name" : "Car Share 1", "metres" : 2000 }}}'

List carShares:
	curl -X GET http://localhost:31415/v0/carShares

List paginated carShares:
	curl -X GET 'http://localhost:31415/v0/carShares?page\[offset\]=0&page\[limit\]=2'
OR
	curl -X GET 'http://localhost:31415/v0/carShares?page\[number\]=1&page\[size\]=2'

Update:
	curl -vX PATCH http://localhost:31415/v0/carShares/1 -d '{ "data" : {"type" : "carShares", "id": "1", "attributes": {"metres-as-driver" : 1, "metres-as-passenger" : 2}}}'

Delete:
	curl -vX DELETE http://localhost:31415/v0/carShares/1
*/
package main

import (
	"fmt"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resolver"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage"
	"github.com/julienschmidt/httprouter"
	"github.com/manyminds/api2go"
)

func main() {
	port := 31415
	api := api2go.NewAPIWithResolver("v0", &resolver.RequestURL{Port: port})
	tripStorage := storage.NewTripStorage()
	userStorage := storage.NewUserStorage()
	carShareStorage := storage.NewCarShareStorage()
	api.AddResource(model.User{}, resource.UserResource{UserStorage: userStorage})
	api.AddResource(model.Trip{}, resource.TripResource{TripStorage: tripStorage})
	api.AddResource(model.CarShare{}, resource.CarShareResource{CarShareStorage: carShareStorage, TripStorage: tripStorage, UserStorage: userStorage})

	fmt.Printf("Listening on :%d", port)
	handler := api.Handler().(*httprouter.Router)

	http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}
