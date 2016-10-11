/*
RESTful API for the car share system

Create a new trip:
	curl -X POST http://localhost:31415/v0/trips -d '{"data" : {"type" : "trips" , "attributes": {"km-as-driver" : 0, "km-as-passenger" : 0 }}}'

List tips:
	curl -X GET http://localhost:31415/v0/trips

List paginated tripa:
	curl -X GET 'http://localhost:31415/v0/trips?page\[offset\]=0&page\[limit\]=2'
OR
	curl -X GET 'http://localhost:31415/v0/trips?page\[number\]=1&page\[size\]=2'

Update:
	curl -vX PATCH http://localhost:31415/v0/trips/1 -d '{ "data" : {"type" : "trips", "id": "1", "attributes": {"km-as-driver" : 1, "km-as-passenger" : 2}}}'

Delete:
	curl -vX DELETE http://localhost:31415/v0/trips/2
*/
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
	api.AddResource(model.Trip{}, resource.TripResource{TripStorage: tripStorage})

	fmt.Printf("Listening on :%d", port)
	handler := api.Handler().(*httprouter.Router)

	http.ListenAndServe(fmt.Sprintf(":%d", port), handler)
}
