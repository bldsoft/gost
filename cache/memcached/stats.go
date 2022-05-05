package memcached

type Stats struct {
	Instance string `json:"instance" mapstructure:"instance"`

	Uptime       uint32 `json:"uptime" mapstructure:"uptime"`
	Version      string `json:"version" mapstructure:"version"`
	CurrentItems uint32 `json:"currentItems" mapstructure:"curr_items"`
	TotalItems   uint32 `json:"totalItems" mapstructure:"total_items"`
	CurrentBytes uint64 `json:"currentBytes" mapstructure:"bytes"`
	MaxBytes     uint64 `json:"maxBytes" mapstructure:"limit_maxbytes"`

	Hits   uint64 `json:"hits" mapstructure:"get_hits"`
	Misses uint64 `json:"misses" mapstructure:"get_misses"`

	StoreTooLarge uint64 `json:"storeTooLarge,omitempty" mapstructure:"store_too_large"`
	StoreNoMemory uint64 `json:"storeNoMemory,omitempty" mapstructure:"store_no_memory"`

	Evictions        uint64 `json:"evictions" mapstructure:"evictions"`
	ExpiredUnfetched uint64 `json:"expiredUnfetched,omitempty" mapstructure:"expired_unfetched"`
	EvictedUnfetched uint64 `json:"evictedUnfetched,omitempty" mapstructure:"evicted_unfetched"`
	EvictedActive    uint64 `json:"evictedActive,omitempty" mapstructure:"evicted_active"`

	IncrHits   uint64 `json:"incrHits,omitempty" mapstructure:"incr_hits"`
	IncrMisses uint64 `json:"incrMisses,omitempty" mapstructure:"incr_isses"`

	DecrHits   uint64 `json:"decrHits,omitempty" mapstructure:"decr_hits"`
	DecrMisses uint64 `json:"decrMisses,omitempty" mapstructure:"decr_misses"`

	CasHits   uint64 `json:"casHits,omitempty" mapstructure:"cas_hits"`
	CasMisses uint64 `json:"casMisses,omitempty" mapstructure:"cas_misses"`
	CasBadVal uint64 `json:"casBadVal,omitempty" mapstructure:"cas_badval"`

	TouchHits   uint64 `json:"touchHits,omitempty" mapstructure:"touch_hits"`
	TouchMisses uint64 `json:"touchMisses,omitempty" mapstructure:"touch_misses"`

	DeleteHits   uint64 `json:"deleteHits,omitempty" mapstructure:"delete_hits"`
	DeleteMisses uint64 `json:"deleteMisses,omitempty" mapstructure:"delete_misses"`
}
