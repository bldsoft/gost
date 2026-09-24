package utils

import "slices"

func IsIn[T comparable](first T, slice ...T) bool {
	return slices.Contains(slice, first)
}
