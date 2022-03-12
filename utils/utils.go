package utils

import (
	"strconv"
	"time"
	"unsafe"

	"golang.org/x/exp/constraints"
)

func TimeTrack(f func()) (d time.Duration) {
	start := time.Now()
	defer func() { d = time.Since(start) }()
	f()
	return d
}

type Parsed interface {
	constraints.Integer | constraints.Unsigned | constraints.Float | ~bool | ~string
}

func parseInt[T constraints.Signed](s string) (result T, err error) {
	temp, err := strconv.ParseInt(s, 10, int(8*unsafe.Sizeof(result)))
	return T(temp), err
}

func parseUint[T constraints.Unsigned](s string) (result T, err error) {
	temp, err := strconv.ParseUint(s, 10, int(8*unsafe.Sizeof(result)))
	return T(temp), err
}

func Parse[T Parsed](s string) (result T, err error) {
	var ret any
	switch any(result).(type) {
	case bool:
		ret, err = strconv.ParseBool(s)
	case int:
		ret, err = parseInt[int](s)
	case int8:
		ret, err = parseInt[int8](s)
	case int16:
		ret, err = parseInt[int16](s)
	case int32:
		ret, err = parseInt[int32](s)
	case int64:
		ret, err = parseInt[int64](s)
	case uint:
		ret, err = parseUint[uint](s)
	case uint8:
		ret, err = parseUint[uint8](s)
	case uint16:
		ret, err = parseUint[uint16](s)
	case uint32:
		ret, err = parseUint[uint32](s)
	case uint64:
		ret, err = parseUint[uint64](s)
	case uintptr:
		ret, err = parseUint[uintptr](s)
	case float32:
		f, e := strconv.ParseFloat(s, 32)
		ret, err = float32(f), e
	case float64:
		ret, err = strconv.ParseFloat(s, 64)
	case string:
		ret = s
	}
	return ret.(T), err
}
