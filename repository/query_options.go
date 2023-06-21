package repository

type QueryOptions struct {
	Archived bool
	Fields   []string // option for read operations, empty slice means all
	Sort     SortOpt
	Filter   interface{}
}

type SortField struct {
	Field string
	Desc  bool
}

type SortOpt []SortField

func Sort() SortOpt { return nil }

func (o SortOpt) Asc(field string) SortOpt {
	return append(o, SortField{Field: field})
}

func (o SortOpt) Desc(field string) SortOpt {
	return append(o, SortField{Field: field, Desc: true})
}
