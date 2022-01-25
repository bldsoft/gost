package mongo

const (
	bsonFieldNameArchived = "archived"
)

type ArchivedEntity struct {
	Archived bool `json:"archived,omitempty" bson:"archived,omitempty"`
}
