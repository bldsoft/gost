package mongo

const (
	BsonFieldNameArchived = "archived"
)

type EntityArchived struct {
	Archived bool `json:"archived,omitempty" bson:"archived,omitempty"`
}
