package repository

type IEntityID interface {
	RawID() interface{}
	StringID() string
	SetIDFromString(string) error
	GenerateID()
	IsZeroID() bool
}
type IEntityIDPtr[T any] interface {
	*T
	IEntityID
}

func ToRawID[T any, U IEntityIDPtr[T]](id interface{}) interface{} {
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

func ToStringID[T any, U IEntityIDPtr[T]](id interface{}) string {
	switch v := id.(type) {
	case string:
		var e T
		if err := U(&e).SetIDFromString(v); err == nil {
			return U(&e).StringID()
		} else {
			panic("failed to get string id: " + err.Error())
		}
	case IEntityID:
		return v.StringID()
	default:
		panic("failed to get string id: wrong type")
	}
}

func StringsToRawIDs[T any, U IEntityIDPtr[T]](ids []string) []interface{} {
	rawIDs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		rawIDs = append(rawIDs, ToRawID[T, U](id))
	}
	return rawIDs
}

func ToRawIDs[T any, U IEntityIDPtr[T]](ids []interface{}) []interface{} {
	rawIDs := make([]interface{}, 0, len(ids))
	for _, id := range ids {
		rawIDs = append(rawIDs, ToRawID[T, U](id))
	}
	return rawIDs
}

func ToStringIDs[T any, U IEntityIDPtr[T]](ids []interface{}) []string {
	stringIDs := make([]string, 0, len(ids))
	for _, id := range ids {
		stringIDs = append(stringIDs, ToStringID[T, U](id))
	}
	return stringIDs
}
