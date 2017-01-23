package storage

import (
	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// TripStorage interface for trip stores. All trips must be tied to a car share.
type TripStorage interface {

	// Get all trips for a given car share
	GetAll(carShareID string, context api2go.APIContexter) (map[string]model.Trip, error)

	// Get a trip from a car share
	GetOne(carShareID string, id string, context api2go.APIContexter) (model.Trip, error)

	// Insert a trip into a car share
	Insert(carShareID string, t model.Trip, context api2go.APIContexter) (string, error)

	// Delete a trip from a car share
	Delete(carShareID string, id string, context api2go.APIContexter) error

	// Update a trip in a car share
	Update(carShareID string, t model.Trip, context api2go.APIContexter) error

	// Get latest trip in a car share
	GetLatest(carShareID string, context api2go.APIContexter) (model.Trip, error)
}
