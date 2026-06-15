package syncset

import (
	"iter"
	"sync"

	"github.com/bldsoft/gost/utils"
)

// Set is a thread-safe wrapper around utils.Set guarded by an RWMutex.
type Set[K comparable] struct {
	mu  sync.RWMutex
	set utils.Set[K]
}

func NewSet[K comparable]() *Set[K] {
	return &Set[K]{
		set: utils.NewSet[K](),
	}
}

func SetOf[K comparable](vals ...K) *Set[K] {
	s := NewSet[K]()
	s.Put(vals...)
	return s
}

func (s *Set[K]) Put(vals ...K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Put(vals...)
}

func (s *Set[K]) Has(val K) bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Has(val)
}

func (s *Set[K]) Remove(vals ...K) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Remove(vals...)
}

func (s *Set[K]) Clear() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.set.Clear()
}

func (s *Set[K]) Empty() bool {
	return s.Len() == 0
}

func (s *Set[K]) Len() int {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.Len()
}

func (s *Set[K]) Each(fn func(key K)) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	s.set.Each(fn)
}

func (s *Set[K]) ToSlice() []K {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.set.ToSlice()
}

func (s *Set[K]) Iter() iter.Seq[K] {
	return func(yield func(key K) bool) {
		s.mu.RLock()
		defer s.mu.RUnlock()
		s.set.Iter()(yield)
	}
}

func SetUnion[K comparable](a, b *Set[K]) *Set[K] {
	res := NewSet[K]()
	for _, src := range []*Set[K]{a, b} {
		src.mu.RLock()
		src.set.Each(func(k K) {
			res.Put(k)
		})
		src.mu.RUnlock()
	}
	return res
}

func SetDiff[K comparable](a, b *Set[K]) *Set[K] {
	res := NewSet[K]()
	a.mu.RLock()
	keys := a.set.ToSlice()
	a.mu.RUnlock()
	for _, k := range keys {
		if !b.Has(k) {
			res.Put(k)
		}
	}
	return res
}

func SetIntersection[K comparable](a, b *Set[K]) *Set[K] {
	var lenA, lenB int
	a.mu.RLock()
	lenA = a.set.Len()
	a.mu.RUnlock()
	b.mu.RLock()
	lenB = b.set.Len()
	b.mu.RUnlock()

	var small, large *Set[K]
	if lenA > lenB {
		small, large = b, a
	} else {
		small, large = a, b
	}

	res := NewSet[K]()
	small.mu.RLock()
	keys := small.set.ToSlice()
	small.mu.RUnlock()
	for _, k := range keys {
		if large.Has(k) {
			res.Put(k)
		}
	}
	return res
}
