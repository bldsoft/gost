package mongo

const (
	BsonFieldNameArchived = "archived"
)

type IEntityArchived interface {
	IsArchived() bool
}
type EntityArchived struct {
	Archived bool `json:"archived,omitempty" bson:"archived,omitempty"`
}

func (e EntityArchived) IsArchived() bool {
	return e.Archived
}
