package utils

func IsIn[T comparable](first T, slice ...T) bool {
	for _, v := range slice {
		if first == v {
			return true
		}
	}
	return false
}
