package changelog

import (
	"context"
	"encoding/json"
	"time"

	"github.com/bldsoft/gost/auth"
	"github.com/bldsoft/gost/repository"
	"github.com/go-chi/chi/v5/middleware"
)

const BsonFieldNameUserID = "userID"
const BsonFieldNameTimestamp = "timestamp"
const BsonFieldNameEntity = "entity"
const BsonFieldNameEntityID = "entityID"
const BsonFieldNameOperation = "operation"
const BsonFieldNameData = "data"

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

type Record struct {
	UserID    string    `json:"userID,omitempty" bson:"userID,omitempty"`
	Timestamp int64     `json:"timestamp" bson:"timestamp"`
	Operation Operation `json:"operation" bson:"operation"`
	Entity    string    `json:"entity" bson:"entity"`
	EntityID  string    `json:"entityID" bson:"entityID"`
	RequestID string    `json:"requestID" bson:"requestID"`
	Data      string    `json:"data" bson:"data"`
}

type RecordFilter struct{}

func NewRecord(ctx context.Context, collectionName string, op Operation, entity repository.IEntityID) (*Record, error) {
	rec := &Record{
		Timestamp: time.Now().Unix(),
		Operation: op,
		Entity:    collectionName,
		RequestID: middleware.GetReqID(ctx),
	}

	user, ok := auth.UserFromContext(ctx).(repository.IEntityID)
	if ok {
		rec.UserID = user.StringID()
	}

	if entity != nil {
		rec.SetData(entity)
		rec.EntityID = entity.StringID()
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
