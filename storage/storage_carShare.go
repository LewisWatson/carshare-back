package storage

import "github.com/LewisWatson/carshare-back/model"

// CarShareStorage stores all car shares
type CarShareStorage interface {
	GetAll() ([]model.CarShare, error)
	GetOne(id string) (model.CarShare, error)
	Insert(c model.CarShare) (string, error)
	Delete(id string) error
	Update(c model.CarShare) error
}
