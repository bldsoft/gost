package routing

import (
	"net"
	"path"

	"github.com/bldsoft/gost/auth/acl"
	"github.com/bldsoft/gost/utils"
	"golang.org/x/exp/constraints"
)

type MatcherAnyOf[T comparable] struct {
	Values []T `json:"args,omitempty" bson:"args,omtempty" label:"values" description:"the values to match against"`
}

func AnyOf[T comparable](args ...T) *MatcherAnyOf[T] {
	return &MatcherAnyOf[T]{Values: args}
}

func (m MatcherAnyOf[T]) MatchValue(val T) (bool, error) {
	switch v := any(val).(type) {
	case string:
		for _, mask := range m.Values {
			if match, err := path.Match(any(mask).(string), v); err != nil || match {
				return match, err
			}
		}
		return false, nil
	default:
		return utils.IsIn(val, m.Values...), nil
	}
}

//=============================================================================

type MatcherNotAnyOf[T comparable] struct {
	Values []T `json:"args,omitempty" bson:"args,omtempty" label:"values" description:"the values to match against"`
}

func NotAnyOf[T comparable](args ...T) *MatcherAnyOf[T] {
	return &MatcherAnyOf[T]{Values: args}
}

func (m MatcherNotAnyOf[T]) MatchValue(val T) (bool, error) {
	switch v := any(val).(type) {
	case string:
		for _, mask := range m.Values {
			if match, err := path.Match(any(mask).(string), v); err != nil || match {
				return !match, err
			}
		}
		return true, nil
	default:
		return !utils.IsIn(val, m.Values...), nil
	}
}

//=============================================================================

type MatcherAnyOfPtr[T comparable] struct {
	Values []T `json:"args,omitempty" bson:"args,omtempty" label:"values" description:"the values to match against"`
}

func AnyOfPtr[T comparable](args ...T) *MatcherAnyOfPtr[T] {
	return &MatcherAnyOfPtr[T]{Values: args}
}

func (m MatcherAnyOfPtr[T]) MatchValue(val *T) (bool, error) {
	switch v := any(val).(type) {
	case *string:
		if v == nil {
			return false, nil
		}
		for _, mask := range m.Values {
			if match, err := path.Match(any(mask).(string), *v); err != nil || match {
				return match, err
			}
		}
		return false, nil
	default:
		return utils.IsIn(*val, m.Values...), nil
	}
}

//=============================================================================

type MatcherNotAnyOfPtr[T comparable] struct {
	Values []T `json:"args,omitempty" bson:"args,omtempty" label:"values"  description:"the values to match against"`
}

func NotAnyOfPtr[T comparable](args ...T) *MatcherNotAnyOfPtr[T] {
	return &MatcherNotAnyOfPtr[T]{Values: args}
}

func (m MatcherNotAnyOfPtr[T]) MatchValue(val *T) (bool, error) {
	switch v := any(val).(type) {
	case *string:
		if v == nil {
			return false, nil
		}
		for _, mask := range m.Values {
			if match, err := path.Match(any(mask).(string), *v); err != nil || match {
				return !match, err
			}
		}
		return true, nil
	default:
		return !utils.IsIn(*val, m.Values...), nil
	}
}

//=============================================================================

type Range[T constraints.Ordered] struct {
	Left, Right T
}

func (r Range[T]) InRange(val T) bool {
	return r.Left <= val && val <= r.Right
}

type MatcherInRange[T constraints.Ordered] struct {
	Range Range[T] `json:"range,omitempty" bson:"range,omtempty" label:"range"`
}

func InRange[T constraints.Ordered](r Range[T]) *MatcherInRange[T] {
	return &MatcherInRange[T]{Range: r}
}

func (m MatcherInRange[T]) MatchValue(val T) (bool, error) {
	return m.Range.InRange(val), nil
}

//=============================================================================

type MatcherNotInRange[T constraints.Ordered] struct {
	Range Range[T] `json:"range,omitempty" bson:"range,omtempty" label:"range"`
}

func NotInRange[T constraints.Ordered](r Range[T]) *MatcherNotInRange[T] {
	return &MatcherNotInRange[T]{Range: r}
}

func (m MatcherNotInRange[T]) MatchValue(val T) (bool, error) {
	return m.Range.InRange(val), nil
}

//=============================================================================

type MatcherZero[T comparable] struct{}

func MatchesZero[T comparable]() *MatcherZero[T] {
	return &MatcherZero[T]{}
}

func (m MatcherZero[T]) MatchValue(val T) (bool, error) {
	var zero T
	return zero == val, nil
}

//=============================================================================

type MatcherNotZero[T comparable] struct{}

func MatchesNotZero[T comparable]() *MatcherNotZero[T] {
	return &MatcherNotZero[T]{}
}

func (m MatcherNotZero[T]) MatchValue(val T) (bool, error) {
	var zero T
	return zero != val, nil
}

//=============================================================================

type MatcherClientIPAnyOf struct {
	ACL acl.IpRange `json:"args,omitempty" bson:"args,omtempty" label:"acl"`
}

func ClientIPAnyOf(r acl.IpRange) *MatcherClientIPAnyOf {
	return &MatcherClientIPAnyOf{ACL: r}
}

func (m MatcherClientIPAnyOf) MatchValue(val net.IP) (bool, error) {
	return m.ACL.Contains(val), nil
}

//=============================================================================

type MatcherClientIPNotAnyOf struct {
	ACL acl.IpRange `json:"args,omitempty" bson:"args,omtempty" label:"acl"`
}

func ClientIPNotAnyOf(r acl.IpRange) *MatcherClientIPAnyOf {
	return &MatcherClientIPAnyOf{ACL: r}
}

func (m MatcherClientIPNotAnyOf) MatchValue(val net.IP) (bool, error) {
	return !m.ACL.Contains(val), nil
}
