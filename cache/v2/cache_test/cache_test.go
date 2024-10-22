package cache_test

import (
	"testing"
	"time"

	"github.com/bldsoft/gost/cache"
	"github.com/bldsoft/gost/cache/bigcache"
	"github.com/bldsoft/gost/cache/ristretto"
	"github.com/stretchr/testify/assert"
)

var (
	key  = "key"
	data = []byte("value")
)

func Test_SetFor(t *testing.T) {
	tests := []struct {
		Name string
		Rep  cache.IExpiringCacheRepository
	}{
		{"bigcache", bigcache.NewExpiringRepository("{}")},
		{"ristretto", ristretto.NewRepository("{}")},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			err := tt.Rep.SetFor(key, data, time.Second)
			assert.NoError(t, err)

			time.Sleep(10 * time.Millisecond)

			value, err := tt.Rep.Get(key)
			assert.NoError(t, err)
			assert.Equal(t, data, value)

			time.Sleep(time.Second)

			_, err = tt.Rep.Get(key)
			assert.ErrorIs(t, err, cache.ErrCacheMiss)
		})
	}
}
