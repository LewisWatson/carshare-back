package storage

import "github.com/LewisWatson/carshare-back/model"

type TripStorage interface {
	GetAll() ([]model.Trip, error)
	GetOne(id string) (model.Trip, error)
	Insert(t model.Trip) (string, error)
	Delete(id string) error
	Update(t model.Trip) error
	GetLatest(carShareID string) (model.Trip, error)
}
