package slices

import "slices"

func FindFunc[T any](slice []T, fn func(T) bool) (_ T, found bool) {
	if idx := slices.IndexFunc(slice, fn); idx != -1 {
		return slice[idx], true
	}
	var zero T
	return zero, false
}
