/*
RESTful API for the car share system
*/
package main

import (
	"fmt"
	"net/http"
	"os"

	"github.com/manyminds/api2go"
	"github.com/manyminds/api2go/routing"
	"github.com/LewisWatson/carshare-back/model"
	"github.com/LewisWatson/carshare-back/resource"
	"github.com/LewisWatson/carshare-back/storage/mongodb"
	"github.com/LewisWatson/firebase-jwt-auth"
	"github.com/alecthomas/kingpin"
	"github.com/benbjohnson/clock"
	"github.com/gin-gonic/gin"
	"github.com/op/go-logging"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	"gopkg.in/mgo.v2"
)

var (
	port              = kingpin.Flag("port", "Set port to bind to").Default("31415").Envar("CARSHARE_PORT").Int()
	mgoURL            = kingpin.Flag("mgoURL", "URL to MongoDB server or seed server(s) for clusters").Default("localhost").Envar("CARSHARE_MGO_URL").URL()
	firebaseProjectID = kingpin.Flag("firebase", "Firebase project to use for authentication").Default("ridesharelogger").Envar("CARSHARE_FIREBASE_PROJECT").String()
	acao              = kingpin.Flag("cors", "Enable HTTP Access Control (CORS) for the specified URI").PlaceHolder("URI").Envar("CARSHARE_CORS_URI").String()

	log    = logging.MustGetLogger("main")
	format = logging.MustStringFormatter(
		`%{color}%{time:2006-01-02T15:04:05.999} %{level:.4s} %{id:03x}%{color:reset} %{message}`,
	)

	userStorage     = &mongodb.UserStorage{}
	carShareStorage = &mongodb.CarShareStorage{}
	tripStorage     = &mongodb.TripStorage{}
)

func init() {

	logging.SetBackend(logging.NewBackendFormatter(logging.NewLogBackend(os.Stderr, "", 0), format))

	kingpin.UsageTemplate(kingpin.CompactUsageTemplate).Version("0.4.0").Author("Lewis Watson")
	kingpin.CommandLine.Help = "API for tracking car shares"
	kingpin.Parse()
}

func main() {

	log.Infof("connecting to mongodb server %s%s", (*mgoURL).Host, (*mgoURL).Path)
	db, err := mgo.Dial((*mgoURL).String())
	if err != nil {
		log.Fatalf("error connecting to mongodb server: %s", err)
	}

	log.Infof("using firebase project \"%s\" for authentication", *firebaseProjectID)
	tokenVerifier, err := fireauth.New(*firebaseProjectID)
	if err != nil {
		log.Fatal(err)
	}

	r := gin.Default()
	api := api2go.NewAPIWithRouting(
		"v0",
		api2go.NewStaticResolver("/"),
		routing.Gin(r),
	)

	if *acao != "" {
		log.Infof("enabling CORS access for %s", *acao)
	}

	api.UseMiddleware(
		func(c api2go.APIContexter, w http.ResponseWriter, r *http.Request) {
			// ensure the db connection is always available in the context
			c.Set("db", db)
			if *acao != "" {
				w.Header().Set("Access-Control-Allow-Origin", *acao)
				w.Header().Set("Access-Control-Allow-Headers", "Authorization,content-type")
				w.Header().Set("Access-Control-Allow-Methods", "GET,PATCH,DELETE,OPTIONS")
			}
		},
	)

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

	// handler for metrics
	r.GET("/metrics", gin.WrapH(promhttp.Handler()))

	log.Infof("Listening and serving HTTP on :%d", *port)
	log.Fatal(r.Run(fmt.Sprintf(":%d", *port)))
}
