package seq

import "iter"

func Concat[K any](iterators ...iter.Seq[K]) iter.Seq[K] {
	return func(yield func(k K) bool) {
		for _, iter := range iterators {
			for v := range iter {
				if !yield(v) {
					return
				}
			}
		}
	}
}

func Concat2[K, V any](iterators ...iter.Seq2[K, V]) iter.Seq2[K, V] {
	return func(yield func(k K, v V) bool) {
		for _, iter := range iterators {
			for k, v := range iter {
				if !yield(k, v) {
					return
				}
			}
		}
	}
}
