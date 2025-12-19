package mongo

import "time"

const (
	BsonFieldNameDeleteTime = "deleteTime"
	BsonFieldNameArchived   = "archived"
)

type IEntityArchived interface {
	IsArchived() bool
}
type EntityArchived struct {
	DeleteTime time.Time `json:"deleteTime,omitzero" bson:"deleteTime,omitempty"`
	Archived   bool      `json:"archived,omitempty" bson:"archived,omitempty"` // Backwards compatibility
}

func (e EntityArchived) IsArchived() bool {
	return !e.DeleteTime.IsZero() || e.Archived
}
