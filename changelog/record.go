package changelog

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/bldsoft/gost/utils"
	"github.com/go-chi/chi/v5/middleware"
)

// TODO: move to auth/user package
var UserEntryCtxKey = &utils.ContextKey{"UserEntry"}
var UserNotFound = errors.New("User isn't found in context")

const BsonFieldNameUserID = "userID"
const BsonFieldNameTimestamp = "timestamp"
const BsonFieldNameEntity = "entity"
const BsonFieldNameEntityID = "entityID"

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

type idType = interface{}

type EntityID interface {
	GetID() interface{}
}

type Record struct {
	UserID    idType    `json:"userID" bson:"userID"`
	Timestamp int64     `json:"timestamp" bson:"timestamp"`
	Operation Operation `json:"operation" bson:"operation"`
	Entity    string    `json:"entity" bson:"entity"`
	EntityID  idType    `json:"entityID" bson:"entityID"`
	RequestID string    `json:"requestID" bson:"requestID"`
	Data      string    `json:"data" bson:"data"`
}

func NewRecord(ctx context.Context, collectionName string, op Operation, entity EntityID) (*Record, error) {
	user, ok := ctx.Value(UserEntryCtxKey).(EntityID)
	if !ok {
		return nil, UserNotFound
	}

	rec := &Record{
		UserID:    user.GetID(),
		Timestamp: time.Now().Unix(),
		Operation: op,
		Entity:    collectionName,
		RequestID: middleware.GetReqID(ctx),
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
