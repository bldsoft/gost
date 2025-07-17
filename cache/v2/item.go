package cache

import "time"

type Item struct {
	Value []byte
	TTL   time.Duration
	Flags uint32
}

type ItemF func(*Item)

func CollectItem(itemFs ...ItemF) *Item {
	it := &Item{}
	for _, f := range itemFs {
		f(it)
	}
	return it
}

func WithTTL(ttl time.Duration) ItemF {
	return func(it *Item) {
		it.TTL = ttl
	}
}

func WithFlags(flags uint32) ItemF {
	return func(it *Item) {
		it.Flags = flags
	}
}
