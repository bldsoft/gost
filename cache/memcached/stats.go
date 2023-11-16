package memcached

type Stats struct {
	Instance string `json:"instance" mapstructure:"instance"`

	Version      string `json:"version" mapstructure:"version"`
	CurrentBytes uint64 `json:"currentBytes" mapstructure:"bytes"`
	MaxBytes     uint64 `json:"maxBytes" mapstructure:"limit_maxbytes"`

	Hits   uint64 `json:"hits" mapstructure:"get_hits"`
	Misses uint64 `json:"misses" mapstructure:"get_misses"`

	StoreTooLarge uint64 `json:"storeTooLarge" mapstructure:"store_too_large"`
	StoreNoMemory uint64 `json:"storeNoMemory" mapstructure:"store_no_memory"`

	Evictions        uint64 `json:"evictions" mapstructure:"evictions"`
	ExpiredUnfetched uint64 `json:"expiredUnfetched" mapstructure:"expired_unfetched"`
	EvictedUnfetched uint64 `json:"evictedUnfetched" mapstructure:"evicted_unfetched"`
	EvictedActive    uint64 `json:"evictedActive" mapstructure:"evicted_active"`

	IncrHits   uint64 `json:"incrHits" mapstructure:"incr_hits"`
	IncrMisses uint64 `json:"incrMisses" mapstructure:"incr_isses"`

	DecrHits   uint64 `json:"decrHits" mapstructure:"decr_hits"`
	DecrMisses uint64 `json:"decrMisses" mapstructure:"decr_misses"`

	CasHits   uint64 `json:"casHits" mapstructure:"cas_hits"`
	CasMisses uint64 `json:"casMisses" mapstructure:"cas_misses"`
	CasBadVal uint64 `json:"casBadVal" mapstructure:"cas_badval"`

	TouchHits   uint64 `json:"touchHits" mapstructure:"touch_hits"`
	TouchMisses uint64 `json:"touchMisses" mapstructure:"touch_misses"`

	DeleteHits   uint64 `json:"deleteHits" mapstructure:"delete_hits"`
	DeleteMisses uint64 `json:"deleteMisses" mapstructure:"delete_misses"`

	Uptime       uint32 `json:"uptime" mapstructure:"uptime"`
	CurrentItems uint32 `json:"currentItems" mapstructure:"curr_items"`
	TotalItems   uint32 `json:"totalItems" mapstructure:"total_items"`
}
