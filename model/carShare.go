package model

// A user of the system
type CarShare struct {
	ID     string  `json:"-"`
	Name   string  `json:"name"`
	Metres int     `json:"metres"`
	Trips  []*Trip `json:"trips"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (cs CarShare) GetID() string {
	return cs.ID
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (cs *CarShare) SetID(id string) error {
	cs.ID = id
	return nil
}
