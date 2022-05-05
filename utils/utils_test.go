package utils

import (
	"fmt"
	"math"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

type test interface {
	name() string
	check(t *testing.T)
}

type parseTest[T Parsed] struct {
	s           string
	expectedRes T
	errExpected bool
}

func positiveTest[T Parsed](s string, expectedRes T) *parseTest[T] {
	return &parseTest[T]{s, expectedRes, false}
}
func negativeTest[T Parsed](s string) *parseTest[T] {
	return &parseTest[T]{s: s, errExpected: true}
}

func (test *parseTest[T]) check(t *testing.T) {
	res, err := Parse[T](test.s)
	if test.errExpected {
		assert.Error(t, err)
	} else {
		assert.NoError(t, err)
		assert.Equal(t, test.expectedRes, res)
	}
}

func (test *parseTest[T]) name() string {
	return fmt.Sprintf("Parse[%T](%s)", test.expectedRes, test.s)
}

func TestParse(t *testing.T) {
	tests := []test{
		positiveTest("1", 1),
		positiveTest("0", 0),
		positiveTest("-1", -1),
		positiveTest("0", uint(0)),

		negativeTest[uint]("-1"),
		negativeTest[uint8]("-1"),
		negativeTest[uint16]("-1"),
		negativeTest[uint32]("-1"),
		negativeTest[uint64]("-1"),

		negativeTest[uint]("A"),
		negativeTest[int]("A"),

		positiveTest(strconv.FormatInt(math.MinInt, 10), int(math.MinInt)),
		positiveTest(strconv.FormatInt(math.MinInt8, 10), int8(math.MinInt8)),
		positiveTest(strconv.FormatInt(math.MinInt16, 10), int16(math.MinInt16)),
		positiveTest(strconv.FormatInt(math.MinInt32, 10), int32(math.MinInt32)),
		positiveTest(strconv.FormatInt(math.MinInt64, 10), int64(math.MinInt64)),

		positiveTest(strconv.FormatInt(math.MaxInt, 10), int(math.MaxInt)),
		positiveTest(strconv.FormatInt(math.MaxInt8, 10), int8(math.MaxInt8)),
		positiveTest(strconv.FormatInt(math.MaxInt16, 10), int16(math.MaxInt16)),
		positiveTest(strconv.FormatInt(math.MaxInt32, 10), int32(math.MaxInt32)),
		positiveTest(strconv.FormatInt(math.MaxInt64, 10), int64(math.MaxInt64)),
		positiveTest(strconv.FormatUint(math.MaxUint, 10), uint(math.MaxUint)),
		positiveTest(strconv.FormatUint(math.MaxUint8, 10), uint8(math.MaxUint8)),
		positiveTest(strconv.FormatUint(math.MaxUint16, 10), uint16(math.MaxUint16)),
		positiveTest(strconv.FormatUint(math.MaxUint32, 10), uint32(math.MaxUint32)),
		positiveTest(strconv.FormatUint(math.MaxUint64, 10), uint64(math.MaxUint64)),

		//overflow
		negativeTest[int8]("-129"),
		negativeTest[int16]("-32769"),
		negativeTest[int32]("-2147483649"),
		negativeTest[int64]("-9223372036854775809"),

		negativeTest[int8]("128"),
		negativeTest[int16]("32768"),
		negativeTest[int32]("2147483648"),
		negativeTest[int64]("9223372036854775808"),

		negativeTest[uint8]("256"),
		negativeTest[uint16]("65536"),
		negativeTest[uint32]("4294967296"),
		negativeTest[uint64]("18446744073709551616"),
		//=====

		positiveTest("0.0", 0.0),
		positiveTest("1", 1.0),

		positiveTest("string", "string"),

		positiveTest("TRUE", true),
		positiveTest("True", true),
		positiveTest("true", true),
		positiveTest("T", true),
		positiveTest("t", true),
		positiveTest("1", true),

		positiveTest("FALSE", false),
		positiveTest("False", false),
		positiveTest("false", false),
		positiveTest("F", false),
		positiveTest("f", false),
		positiveTest("0", false),

		negativeTest[bool]("2"),
		negativeTest[bool]("-1"),
	}

	for _, test := range tests {
		t.Run(test.name(), test.check)
	}
}
