package mongo

const (
	bsonFieldNameArchived = "archived"
)

type EntityArchived struct {
	Archived bool `json:"archived,omitempty" bson:"archived,omitempty"`
}
