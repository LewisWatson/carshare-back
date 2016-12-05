package storage

import (
	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

type UserStorage interface {
	GetAll(context api2go.APIContexter) ([]model.User, error)
	GetOne(id string, context api2go.APIContexter) (model.User, error)
	Insert(u model.User, context api2go.APIContexter) (string, error)
	Delete(id string, context api2go.APIContexter) error
	Update(u model.User, context api2go.APIContexter) error
}
