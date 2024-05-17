package utils

import "encoding/json"

type StringHider interface {
	Hide(string) string
}

type starHider struct{}

func (starHider) Hide(string) string {
	return "****"
}

type last4Hider struct{}

func (last4Hider) Hide(s string) string {
	const visibleCnt = 4
	if len(s) < visibleCnt {
		return "****"
	}

	return "****" + string(s[len(s)-visibleCnt:])
}

type HiddenString[T StringHider] string

func (c HiddenString[T]) MarshalJSON() ([]byte, error) {
	var h T
	return json.Marshal(h.Hide(c.String()))
}

func (c HiddenString[T]) String() string {
	return string(c)
}

type FullyHidden = HiddenString[starHider]
type PartiallyHidden = HiddenString[last4Hider]
