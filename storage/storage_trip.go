package storage

import (
	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

type TripStorage interface {
	GetAll(context api2go.APIContexter) ([]model.Trip, error)
	GetOne(id string, context api2go.APIContexter) (model.Trip, error)
	Insert(t model.Trip, context api2go.APIContexter) (string, error)
	Delete(id string, context api2go.APIContexter) error
	Update(t model.Trip, context api2go.APIContexter) error
	GetLatest(carShareID string, context api2go.APIContexter) (model.Trip, error)
}
