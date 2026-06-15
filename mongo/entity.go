package mongo

import (
	"go.mongodb.org/mongo-driver/v2/bson"
)

type EntityID struct {
	ID bson.ObjectID `json:"id" bson:"_id,omitempty"`
}

func (e *EntityID) RawID() interface{} {
	return e.ID
}

func (e *EntityID) SetIDFromString(id string) error {
	objID, err := bson.ObjectIDFromHex(id)
	if err == nil {
		e.ID = objID
	}
	return err
}

func (e *EntityID) GenerateID() {
	e.ID = bson.NewObjectID()
}

func (e *EntityID) StringID() string {
	return e.ID.Hex()
}

func (e *EntityID) IsZeroID() bool {
	return e.ID == bson.NilObjectID
}
