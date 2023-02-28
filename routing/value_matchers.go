package routing

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"path"

	"github.com/bldsoft/gost/auth/acl"
	"github.com/bldsoft/gost/utils"
	"golang.org/x/exp/constraints"
)

type MatcherAnyOf[T comparable] struct {
	Values []T `json:"args,omitempty" bson:"args,omtempty"`
}

func AnyOf[T comparable](args ...T) *MatcherAnyOf[T] {
	return &MatcherAnyOf[T]{Values: args}
}

func (m MatcherAnyOf[T]) Args() []T {
	return m.Values
}

func (m *MatcherAnyOf[T]) SetArgs(args []T) error {
	m.Values = args
	return nil
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
	Values []T `json:"args,omitempty" bson:"args,omtempty"`
}

func NotAnyOf[T comparable](args ...T) *MatcherAnyOf[T] {
	return &MatcherAnyOf[T]{Values: args}
}

func (m MatcherNotAnyOf[T]) Args() []T {
	return m.Values
}

func (m *MatcherNotAnyOf[T]) SetArgs(args []T) error {
	m.Values = args
	return nil
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

type Range[T constraints.Ordered] struct {
	Left, Right T
}

func (r Range[T]) InRange(val T) bool {
	return r.Left <= val && val <= r.Right
}

type MatcherInRange[T constraints.Ordered] struct {
	Range Range[T] `json:"range,omitempty" bson:"range,omtempty"`
}

func InRange[T constraints.Ordered](r Range[T]) *MatcherInRange[T] {
	return &MatcherInRange[T]{Range: r}
}

func (m MatcherInRange[T]) Args() Range[T] {
	return m.Range
}

func (m *MatcherInRange[T]) SetArgs(args Range[T]) error {
	m.Range = args
	return nil
}

func (m MatcherInRange[T]) MatchValue(val T) (bool, error) {
	return m.Range.InRange(val), nil
}

//=============================================================================

type MatcherNotInRange[T constraints.Ordered] struct {
	Range Range[T] `json:"range,omitempty" bson:"range,omtempty"`
}

func NotInRange[T constraints.Ordered](r Range[T]) *MatcherNotInRange[T] {
	return &MatcherNotInRange[T]{Range: r}
}

func (m MatcherNotInRange[T]) Args() Range[T] {
	return m.Range
}

func (m *MatcherNotInRange[T]) SetArgs(args Range[T]) error {
	m.Range = args
	return nil
}

func (m MatcherNotInRange[T]) MatchValue(val T) (bool, error) {
	return m.Range.InRange(val), nil
}

//=============================================================================

type QueryOrHeader = interface{ http.Header | url.Values }
type MatcherQueryOrHeaderAnyOf[T QueryOrHeader] struct {
	Values []string `json:"args,omitempty" bson:"args,omtempty"`
}

func MatchesQueryAnyOf[T QueryOrHeader](name string, args ...string) *MatcherQueryOrHeaderAnyOf[T] {
	return &MatcherQueryOrHeaderAnyOf[T]{Values: append([]string{name}, args...)}
}

func (m MatcherQueryOrHeaderAnyOf[T]) Args() []string {
	return m.Values
}

func (m *MatcherQueryOrHeaderAnyOf[T]) SetArgs(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("args must have at least 2 elements")
	}
	m.Values = args
	return nil
}

func (m MatcherQueryOrHeaderAnyOf[T]) MatchValue(val T) (bool, error) {
	for _, mask := range m.Values[1:] {
		if match, err := path.Match(mask, m.Values[0]); err != nil || match {
			return match, err
		}
	}
	return false, nil
}

//=============================================================================

type MatcherQueryOrHeaderNotAnyOf[T QueryOrHeader] struct {
	Values []string `json:"args,omitempty" bson:"args,omtempty"`
}

func MatchesQueryNotAnyOf[T QueryOrHeader](name string, args ...string) *MatcherQueryOrHeaderNotAnyOf[T] {
	return &MatcherQueryOrHeaderNotAnyOf[T]{Values: append([]string{name}, args...)}
}

func (m MatcherQueryOrHeaderNotAnyOf[T]) Args() []string {
	return m.Values
}

func (m *MatcherQueryOrHeaderNotAnyOf[T]) SetArgs(args []string) error {
	if len(args) < 2 {
		return fmt.Errorf("args must have at least 2 elements")
	}
	m.Values = args
	return nil
}

func (m MatcherQueryOrHeaderNotAnyOf[T]) MatchValue(val T) (bool, error) {
	for _, mask := range m.Values[1:] {
		if match, err := path.Match(mask, m.Values[0]); err != nil || match {
			return !match, err
		}
	}
	return false, nil
}

//=============================================================================

type MatcherQueryOrHeaderExists[T QueryOrHeader] struct {
	Name string `json:"args,omitempty" bson:"args,omtempty"`
}

func MatchesQueryOrHeaderExists[T QueryOrHeader](name string) *MatcherQueryOrHeaderExists[T] {
	return &MatcherQueryOrHeaderExists[T]{Name: name}
}

func (m MatcherQueryOrHeaderExists[T]) Args() string {
	return m.Name
}

func (m *MatcherQueryOrHeaderExists[T]) SetArgs(args string) error {
	m.Name = args
	return nil
}

func (m MatcherQueryOrHeaderExists[T]) MatchValue(val T) (bool, error) {
	_, ok := val[m.Name]
	return ok, nil
}

//=============================================================================

type MatcherQueryOrHeaderNotExists[T QueryOrHeader] struct {
	Name string `json:"args,omitempty" bson:"args,omtempty"`
}

func MatchesNotExists[T QueryOrHeader](name string) *MatcherQueryOrHeaderNotExists[T] {
	return &MatcherQueryOrHeaderNotExists[T]{Name: name}
}

func (m MatcherQueryOrHeaderNotExists[T]) Args() string {
	return m.Name
}

func (m *MatcherQueryOrHeaderNotExists[T]) SetArgs(args string) error {
	m.Name = args
	return nil
}

func (m MatcherQueryOrHeaderNotExists[T]) MatchValue(val T) (bool, error) {
	_, ok := val[m.Name]
	return !ok, nil
}

//=============================================================================

type MatcherClientIPAnyOf struct {
	ACL acl.IpRange `json:"args,omitempty" bson:"args,omtempty"`
}

func ClientIPAnyOf(r acl.IpRange) *MatcherClientIPAnyOf {
	return &MatcherClientIPAnyOf{ACL: r}
}

func (m MatcherClientIPAnyOf) Args() acl.IpRange {
	return m.ACL
}

func (m *MatcherClientIPAnyOf) SetArgs(r acl.IpRange) error {
	m.ACL = r
	return nil
}

func (m MatcherClientIPAnyOf) MatchValue(val net.IP) (bool, error) {
	return m.ACL.Contains(val), nil
}

//=============================================================================

type MatcherClientIPNotAnyOf struct {
	ACL acl.IpRange `json:"args,omitempty" bson:"args,omtempty"`
}

func ClientIPNotAnyOf(r acl.IpRange) *MatcherClientIPAnyOf {
	return &MatcherClientIPAnyOf{ACL: r}
}

func (m MatcherClientIPNotAnyOf) Args() acl.IpRange {
	return m.ACL
}

func (m *MatcherClientIPNotAnyOf) SetArgs(r acl.IpRange) error {
	m.ACL = r
	return nil
}

func (m MatcherClientIPNotAnyOf) MatchValue(val net.IP) (bool, error) {
	return !m.ACL.Contains(val), nil
}
