package bigcache

import "github.com/bldsoft/gost/cache"

func NewExpiringRepository(jsonConfig string) *cache.ExpiringCacheRepository {
	return cache.NewExpiringRepository(NewRepository(jsonConfig))
}
