package repository

//go:generate go run github.com/dmarkham/enumer@latest -gqlgen -type SortOrder --trimprefix "SortOrder" --output sort_order_enum.go

type SortOrder int

const (
	SortOrderASC SortOrder = iota
	SortOrderDESC
)
