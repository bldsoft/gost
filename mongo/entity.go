package mongo

import (
	"go.mongodb.org/mongo-driver/bson/primitive"
)

type EntityID struct {
	ID primitive.ObjectID `json:"id" bson:"_id,omitempty"`
}

func (e *EntityID) RawID() interface{} {
	return e.ID
}

func (e *EntityID) SetIDFromString(id string) error {
	objID, err := primitive.ObjectIDFromHex(id)
	if err == nil {
		e.ID = objID
	}
	return err
}

func (e *EntityID) GenerateID() {
	e.ID = primitive.NewObjectID()
}

func (e *EntityID) StringID() string {
	return e.ID.Hex()
}

func (e *EntityID) IsZeroID() bool {
	return e.ID == primitive.NilObjectID
}
