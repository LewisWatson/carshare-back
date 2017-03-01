package storage

import (
	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// CarShareStorage stores all car shares
type CarShareStorage interface {
	GetAll(userID string, context api2go.APIContexter) ([]model.CarShare, error)
	GetOne(id string, context api2go.APIContexter) (model.CarShare, error)
	Insert(c model.CarShare, context api2go.APIContexter) (string, error)
	Delete(id string, context api2go.APIContexter) error
	Update(c model.CarShare, context api2go.APIContexter) error
}
