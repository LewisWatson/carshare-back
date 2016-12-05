package model

// A score keeps track of how many miles a user has travelled as a driver and as a passenger
type Score struct {
	MetresAsDriver    int `json:"metres-as-driver"    bson:"metres-as-driver"`
	MetresAsPassenger int `json:"metres-as-passenger" bson:"metres-as-passenger"`
}
