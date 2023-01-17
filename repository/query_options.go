package repository

type QueryOptions struct {
	Archived bool
	Fields   []string // option for read operations, empty slice means all
	Filter   interface{}
	Sort     interface{}
	Skip     int64
	Limit    int64
}
