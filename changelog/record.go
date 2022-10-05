package changelog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bldsoft/gost/auth"
	"github.com/go-chi/chi/v5/middleware"
)

const BsonFieldNameUserID = "userID"
const BsonFieldNameTimestamp = "timestamp"
const BsonFieldNameEntity = "entity"
const BsonFieldNameEntityID = "entityID"
const BsonFieldNameOperation = "operation"

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

func (op Operation) String() string {
	switch op {
	case Create:
		return "CREATE"
	case Update:
		return "UPDATE"
	case Delete:
		return "DELETE"
	default:
		return "NONE"
	}
}

type idType = interface{}

type EntityID interface {
	GetID() interface{}
}

type Record struct {
	UserID    idType    `json:"userID,omitempty" bson:"userID,omitempty"`
	Timestamp int64     `json:"timestamp" bson:"timestamp"`
	Operation Operation `json:"operation" bson:"operation"`
	Entity    string    `json:"entity" bson:"entity"`
	EntityID  idType    `json:"entityID" bson:"entityID"`
	RequestID string    `json:"requestID" bson:"requestID"`
	Data      string    `json:"data" bson:"data"`
}

func NewRecord(ctx context.Context, collectionName string, op Operation, entity EntityID) (*Record, error) {
	rec := &Record{
		Timestamp: time.Now().Unix(),
		Operation: op,
		Entity:    collectionName,
		RequestID: middleware.GetReqID(ctx),
	}

	user, ok := auth.UserFromContext(ctx).(EntityID)
	if ok {
		rec.UserID = user.GetID()
	}

	if entity != nil {
		rec.SetData(entity)
		rec.EntityID = entity.GetID()
	}

	return rec, nil
}

func (r *Record) SetData(entity interface{}) error {
	data, err := json.Marshal(entity)
	if err != nil {
		return err
	}
	r.Data = string(data)
	return nil
}
