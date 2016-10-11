package model

// A trip is a single instance of a car share
type Trip struct {
	ID            string `json:"-"`
	KMAsDriver    int    `json:"km-as-driver"`
	KMAsPassenger int    `json:"km-as-passenger"`
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
