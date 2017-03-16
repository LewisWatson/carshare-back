package storage

import (
	"github.com/LewisWatson/carshare-back/model"
	"github.com/manyminds/api2go"
)

// UserInserter functions related to inserting users into a data store
type UserInserter interface {
	Insert(u model.User, context api2go.APIContexter) (string, error)
}

// UserGetter functions related to retrieving users from a data store
type UserGetter interface {
	GetAll(context api2go.APIContexter) ([]model.User, error)
	GetOne(id string, context api2go.APIContexter) (model.User, error)
}

// UserUpdater functions related to updating users in a data store
type UserUpdater interface {
	Update(u model.User, context api2go.APIContexter) error
}

// UserDeleter functions related to deleting users from a data store
type UserDeleter interface {
	Delete(id string, context api2go.APIContexter) error
}

// FirebaseUserGetter functions related to users linked for firebase users
type FirebaseUserGetter interface {
	GetByFirebaseUID(firebaseUID string, context api2go.APIContexter) (model.User, error)
}

// UserStorage functions for interacting with users in a data store
type UserStorage interface {
	UserInserter
	UserGetter
	UserUpdater
	UserDeleter
	FirebaseUserGetter
}
