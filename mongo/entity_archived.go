package mongo

import "time"

const (
	BsonFieldNameDeletedAt = "deleted_at"
	BsonFieldNameArchived  = "archived"
)

type IEntityArchived interface {
	IsArchived() bool
}
type EntityArchived struct {
	DeletedAt time.Time `json:"deleted_at,omitempty" bson:"deleted_at,omitempty"`
	Archived  bool      `json:"archived,omitempty" bson:"archived,omitempty"`
}

func (e EntityArchived) IsArchived() bool {
	return !e.DeletedAt.IsZero() || e.Archived
}
