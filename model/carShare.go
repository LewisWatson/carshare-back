package model

import (
	"errors"
	"fmt"
	"log"
	"sort"

	"github.com/manyminds/api2go/jsonapi"
	"gopkg.in/mgo.v2/bson"
)

// CarShare an individual group of users who make up a car share
type CarShare struct {
	ID        bson.ObjectId `json:"-"    bson:"_id,omitempty"`
	Name      string        `json:"name" bson:"name"`
	Members   []*User       `json:"-",   bson:"-"`
	MemberIDs []string      `json:"-",   bson:"members"`
	Admins    []*User       `json:"-"    bson:"-"`
	AdminIDs  []string      `json:"-"    bson:"admins"`
	Trips     []Trip        `json:"-"    bson:"-"`
	TripIDs   []string      `json:"-"    bson:"trips"`
}

// GetID to satisfy jsonapi.MarshalIdentifier interface
func (cs CarShare) GetID() string {
	return cs.ID.Hex()
}

// SetID to satisfy jsonapi.UnmarshalIdentifier interface
func (cs *CarShare) SetID(id string) error {
	// for some reason SetID gets called with null ("") when run in TravisCI
	// this doesn't seem to happen during local builds.
	if id == "" {
		return nil
	}
	if bson.IsObjectIdHex(id) {
		cs.ID = bson.ObjectIdHex(id)
		return nil
	}
	return errors.New("<id>" + id + "</id> is not a valid car share id")
}

// GetReferences to satisfy the jsonapi.MarshalReferences interface
func (cs CarShare) GetReferences() []jsonapi.Reference {
	return []jsonapi.Reference{
		{
			Type: "trips",
			Name: "trips",
		},
		{
			Type: "users",
			Name: "members",
		},
		{
			Type: "users",
			Name: "admins",
		},
	}
}

// GetReferencedIDs to satisfy the jsonapi.MarshalLinkedRelations interface
func (cs CarShare) GetReferencedIDs() []jsonapi.ReferenceID {
	result := []jsonapi.ReferenceID{}
	for _, trip := range cs.Trips {
		result = append(result, jsonapi.ReferenceID{
			ID:   trip.GetID(),
			Type: "trips",
			Name: "trips",
		})
	}
	for _, member := range cs.Members {
		result = append(result, jsonapi.ReferenceID{
			ID:   member.GetID(),
			Type: "users",
			Name: "members",
		})
	}
	for _, admin := range cs.Admins {
		result = append(result, jsonapi.ReferenceID{
			ID:   admin.GetID(),
			Type: "users",
			Name: "admins",
		})
	}
	log.Printf("car share %s returning referenced ids %+v", cs.GetID(), result)
	return result
}

// GetReferencedStructs to satisfy the jsonapi.MarhsalIncludedRelations interface
func (cs CarShare) GetReferencedStructs() []jsonapi.MarshalIdentifier {
	result := []jsonapi.MarshalIdentifier{}
	for _, trip := range cs.Trips {
		result = append(result, trip)
	}
	for _, member := range cs.Members {
		result = append(result, member)
	}
	for _, admin := range cs.Admins {
		result = append(result, admin)
	}
	log.Printf("car share %s returning referenced structs %+v", cs.GetID(), result)
	return result
}

// SetToManyReferenceIDs sets the trips reference IDs and satisfies the jsonapi.UnmarshalToManyRelations interface
func (cs *CarShare) SetToManyReferenceIDs(name string, IDs []string) error {
	log.Printf("car share %s setting %s ids %v", cs.GetID(), name, IDs)
	switch name {
	case "trips":
		cs.TripIDs = IDs
		sort.Strings(cs.TripIDs)
		break
	case "members":
		cs.MemberIDs = IDs
		sort.Strings(cs.MemberIDs)
		break
	case "admins":
		cs.AdminIDs = IDs
		sort.Strings(cs.AdminIDs)
		break
	default:
		return fmt.Errorf("There is no to-many relationship with the name " + name)
	}
	return nil
}

// AddToManyIDs adds some new trips
func (cs *CarShare) AddToManyIDs(name string, IDs []string) error {
	log.Printf("car share %s add %s ids, %v", cs.GetID(), name, IDs)
	switch name {
	case "trips":
		cs.TripIDs = append(cs.TripIDs, IDs...)
		sort.Strings(cs.TripIDs)
		break
	case "members":
		cs.MemberIDs = append(cs.MemberIDs, IDs...)
		sort.Strings(cs.MemberIDs)
		break
	case "admins":
		cs.AdminIDs = append(cs.AdminIDs, IDs...)
		sort.Strings(cs.AdminIDs)
		break
	default:
		return errors.New("There is no to-many relationship with the name " + name)
	}
	return nil
}

// DeleteToManyIDs removes some relationships from car shrae
func (cs *CarShare) DeleteToManyIDs(name string, IDs []string) error {
	log.Printf("car share %s remove %s ids, %v", cs.GetID(), name, IDs)
	switch name {
	case "trips":
		for _, ID := range IDs {
			for pos, oldID := range cs.TripIDs {
				if ID == oldID {
					cs.TripIDs = append(cs.TripIDs[:pos], cs.TripIDs[pos+1:]...)
					break
				}
			}
			log.Printf("car share %s unable to find trip %s", cs.GetID(), ID)
		}
		break
	case "members":
		for _, ID := range IDs {
			for pos, oldID := range cs.MemberIDs {
				if ID == oldID {
					cs.AdminIDs = append(cs.MemberIDs[:pos], cs.MemberIDs[pos+1:]...)
					break
				}
			}
			log.Printf("car share %s unable to find member %s", cs.GetID(), ID)
		}
		break
	case "admins":
		for _, ID := range IDs {
			for pos, oldID := range cs.AdminIDs {
				if ID == oldID {
					cs.AdminIDs = append(cs.AdminIDs[:pos], cs.AdminIDs[pos+1:]...)
					break
				}
			}
			log.Printf("car share %s unable to find admin %s", cs.GetID(), ID)
		}
		break
	default:
		return errors.New("There is no to-many relationship with the name " + name)
	}
	return nil
}
