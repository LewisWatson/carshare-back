package model

// A score keeps track of how many miles a user has travelled as a driver and as a passenger
type Score struct {
	UserId            string `json:"user"`
	MetersAsDriver    int    `json:"meters-as-driver"`
	MetersAsPassenger int    `json:"meters-as-passenger"`
}
