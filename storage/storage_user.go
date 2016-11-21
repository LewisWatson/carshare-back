package storage

import "github.com/LewisWatson/carshare-back/model"

type UserStorage interface {
	GetAll() ([]model.User, error)
	GetOne(id string) (model.User, error)
	Insert(u model.User) (string, error)
	Delete(id string) error
	Update(u model.User) error
}
