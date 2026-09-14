package repository

import "fmt"

type IIDProvider interface {
	RawID() any
	StringID() string
	IsZeroID() bool
}

type IEntityID interface {
	IIDProvider
	SetIDFromString(string) error
	GenerateID()
}
type IEntityIDPtr[T any] interface {
	*T
	IEntityID
}

func ToRawID[T any, U IEntityIDPtr[T]](id any) any {
	switch v := id.(type) {
	case string:
		var e T
		if err := U(&e).SetIDFromString(v); err == nil {
			return U(&e).RawID()
		}
		return v
	case IEntityID:
		return v.RawID()
	default:
		return id
	}
}

func ToStringID[T any, U IEntityIDPtr[T]](id any) string {
	switch v := id.(type) {
	case string:
		var e T
		if err := U(&e).SetIDFromString(v); err == nil {
			return U(&e).StringID()
		} else {
			panic(fmt.Sprintf("failed to get string id: %v"+err.Error(), id))
		}
	case IEntityID:
		return v.StringID()
	default:
		panic("failed to get string id: wrong type")
	}
}

func StringsToRawIDs[T any, U IEntityIDPtr[T]](ids []string) []any {
	rawIDs := make([]any, 0, len(ids))
	for _, id := range ids {
		rawIDs = append(rawIDs, ToRawID[T, U](id))
	}
	return rawIDs
}

func ToRawIDs[T any, U IEntityIDPtr[T]](ids []any) []any {
	rawIDs := make([]any, 0, len(ids))
	for _, id := range ids {
		rawIDs = append(rawIDs, ToRawID[T, U](id))
	}
	return rawIDs
}

func ToStringIDs[T any, U IEntityIDPtr[T]](ids []any) []string {
	stringIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		stringIDs = append(stringIDs, ToStringID[T, U](id))
	}
	return stringIDs
}
