/*
RESTful API for the car share system
*/
package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage/mongodb"
	"github.com/benbjohnson/clock"
	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go-adapter/gingonic"

	"gopkg.in/LewisWatson/firebase-jwt-auth.v1"
	"gopkg.in/alecthomas/kingpin.v2"
	"gopkg.in/gin-gonic/gin.v1"
	"gopkg.in/mgo.v2"
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

	log.Printf("connecting to mongodb server %s%s", (*mgoURL).Host, (*mgoURL).Path)
	db, err := mgo.Dial((*mgoURL).String())
	if err != nil {
		log.Fatalf("error connecting to mongodb server: %s", err)
	}

	userStorage := &mongodb.UserStorage{}
	carShareStorage := &mongodb.CarShareStorage{}
	tripStorage := &mongodb.TripStorage{}

	tokenVerifier, err := fireauth.New("ridesharelogger")
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	api := api2go.NewAPIWithRouting(
		"v0",
		api2go.NewStaticResolver("/"),
		gingonic.New(r),
	)

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

	r.GET("/ping", func(c *gin.Context) {
		c.String(200, "pong")
	})

	log.Printf("listening on :%d", *port)
	log.Fatal(r.Run(fmt.Sprintf(":%d", *port)))
}
