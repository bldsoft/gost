package changelog

const BsonFieldNameEntityID = "entityID"
const BsonFieldNameTimestamp = "timestamp"
const BsonFieldNameEntity = "entity"

type Operation int

const (
	Create Operation = iota
	Update
	Delete
)

type idType = interface{}

type Record struct {
	UserID    idType    `json:"userID" bson:"userID"`
	Timestamp int64     `json:"timestamp" bson:"timestamp"`
	Operation Operation `json:"operation" bson:"operation"`
	Entity    string    `json:"entity" bson:"entity"`
	EntityID  idType    `json:"entityID" bson:"entityID"`
	RequestID string    `json:"requestID" bson:"requestID"`
	Data      string    `json:"data" bson:"data"`
}
