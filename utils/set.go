package utils

import (
	"iter"
	"maps"
	"slices"
)

// Set implements a hashset, using the hashmap as the underlying storage.
type Set[K comparable] struct {
	m map[K]struct{}
}

// New returns an empty hashset.
func NewSet[K comparable]() Set[K] {
	return Set[K]{
		m: make(map[K]struct{}),
	}
}

// Of returns a new hashset initialized with the given 'vals'
func SetOf[K comparable](vals ...K) Set[K] {
	s := NewSet[K]()
	for _, val := range vals {
		s.Put(val)
	}
	return s
}

// Put adds 'val' to the set.
func (s Set[K]) Put(vals ...K) {
	for _, val := range vals {
		s.m[val] = struct{}{}
	}
}

// Has returns true only if 'val' is in the set.
func (s Set[K]) Has(val K) bool {
	_, ok := s.m[val]
	return ok
}

// Remove removes 'val' from the set.
func (s Set[K]) Remove(vals ...K) {
	for _, val := range vals {
		delete(s.m, val)
	}
}

// Clear ...
func (s *Set[K]) Clear() {
	*s = NewSet[K]()
}

// Empty ...
func (s Set[K]) Empty() bool {
	return s.Len() == 0
}

// Size returns the number of elements in the set.
func (s Set[K]) Len() int {
	return len(s.m)
}

// Each calls 'fn' on every item in the set in no particular order.
func (s Set[K]) Each(fn func(key K)) {
	for k := range s.m {
		fn(k)
	}
}

func (s Set[K]) ToSlice() []K {
	return slices.Collect(maps.Keys(s.m))
}

func (s Set[K]) Iter() iter.Seq[K] {
	return func(yield func(key K) bool) {
		for k := range s.m {
			if !yield(k) {
				return
			}
		}
	}
}

func SetUnion[K comparable](a, b Set[K]) Set[K] {
	res := NewSet[K]()
	a.Each(func(key K) {
		res.Put(key)
	})
	b.Each(func(key K) {
		res.Put(key)
	})
	return res
}

func SetDiff[K comparable](a, b Set[K]) Set[K] {
	res := NewSet[K]()
	a.Each(func(key K) {
		if !b.Has(key) {
			res.Put(key)
		}
	})
	return res
}

func SetIntersection[K comparable](a, b Set[K]) (intersection Set[K]) {
	if a.Len() > b.Len() {
		a, b = b, a
	}
	res := NewSet[K]()
	a.Each(func(key K) {
		if b.Has(key) {
			res.Put(key)
		}
	})
	return res
}
