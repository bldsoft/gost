package repository

const SchemaTag = "schema"

type QueryOptions[F any] struct {
	Archived bool
	Fields   []string // option for read operations, empty slice means all
	Filter   F
}
