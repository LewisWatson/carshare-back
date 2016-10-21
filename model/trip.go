package model

import "errors"

// A trip is a single instance of a car share
type Trip struct {
	ID     string `json:"-"`
	User   User   `json:"-"`
	UserID string `json:"-"`
	// TimeStamp         time.Time `json:"timestamp"`
	MetersAsDriver    int `json:"meters-as-driver"`
	MetersAsPassenger int `json:"meters-as-passenger"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (t Trip) GetID() string {
	return t.ID
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (t *Trip) SetID(id string) error {
	t.ID = id
	return nil
}

// UnmarshalToOneRelations must be implemented to unmarshal to-one relations
func (t *Trip) SetToOneReferenceID(name, ID string) error {
	if name == "UserId" {
		t.UserID = ID
		return nil
	}

	return errors.New("There is no to-one relationship with the name " + name)
}
