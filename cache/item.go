package cache

import "time"

type Item struct {
	Value    []byte
	TTL      time.Duration
	Flags    uint32
	Deadline time.Time
}

type ItemF func(*Item)

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
