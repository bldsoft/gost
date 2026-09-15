package stat

import (
	"context"
	"errors"
	"strconv"
	"strings"

	aero "github.com/aerospike/aerospike-client-go/v8"
	"github.com/bldsoft/gost/cache/v2/aerospike"
)

const (
	AerospikeKeyPool           = "pool"
	AerospikeKeyMemUtilization = "mem_utilization"
	AerospikeKeyMemUtilOK      = "mem_utilization_ok"
)

func AerospikePayload(pool any, memUtilization float64, memUtilizationOK bool) map[string]any {
	return map[string]any{
		AerospikeKeyPool:           pool,
		AerospikeKeyMemUtilization: memUtilization,
		AerospikeKeyMemUtilOK:      memUtilizationOK,
	}
}

type AerospikeCollector struct {
	cache     *aerospike.Storage
	namespace string
}

func NewAerospikeCollector(cache *aerospike.Storage, namespace string) *AerospikeCollector {
	return &AerospikeCollector{cache: cache, namespace: namespace}
}

func (c *AerospikeCollector) Stat(_ context.Context) Stat {
	if c.cache == nil {
		return NewStat("aerospike", nil, errors.New("aerospike: cache is nil"))
	}

	poolStats, err := c.cache.Stat()
	memU, memOK := 0.0, false
	if err == nil {
		memU, memOK = c.namespaceMaxUtilization()
	}
	pl := AerospikePayload(poolStats, memU, memOK)
	return NewStat("aerospike", pl, err)
}

func (c *AerospikeCollector) namespaceMaxUtilization() (maxUtil float64, valid bool) {
	if c.namespace == "" {
		return 0, false
	}
	nsKey := "namespace/" + c.namespace
	policy := aero.NewInfoPolicy()

	for _, node := range c.cache.GetNodes() {
		if node == nil {
			continue
		}
		infoMap, err := node.RequestInfo(policy, nsKey)
		if err != nil {
			continue
		}
		raw, ok := infoMap[nsKey]
		if !ok || raw == "" {
			continue
		}
		u, ok := namespaceUtilizationFromInfo(raw)
		if !ok {
			continue
		}
		valid = true
		if u > maxUtil {
			maxUtil = u
		}
	}
	return maxUtil, valid
}

func namespaceUtilizationFromInfo(info string) (util float64, ok bool) {
	kv := parseSemicolonKV(info)
	if kv == nil {
		return 0, false
	}
	if kv["stop_writes"] == "true" {
		return 1, true
	}

	var best float64
	found := false
	takeMax := func(u float64, o bool) {
		if !o {
			return
		}
		found = true
		u = clamp01(u)
		if u > best {
			best = u
		}
	}

	if v := kv["memory_free_pct"]; v != "" {
		if pct, err := strconv.ParseFloat(v, 64); err == nil {
			takeMax(1-pct/100, true)
		}
	}
	if v := kv["device_free_pct"]; v != "" {
		if pct, err := strconv.ParseFloat(v, 64); err == nil {
			takeMax(1-pct/100, true)
		}
	}
	if u, o := byteRatio(kv["device_used_bytes"], kv["device_total_bytes"]); o {
		takeMax(u, true)
	}

	return best, found
}

func byteRatio(usedStr, totalStr string) (float64, bool) {
	if usedStr == "" || totalStr == "" {
		return 0, false
	}
	used, err1 := strconv.ParseUint(usedStr, 10, 64)
	total, err2 := strconv.ParseUint(totalStr, 10, 64)
	if err1 != nil || err2 != nil || total == 0 {
		return 0, false
	}
	return float64(used) / float64(total), true
}

func parseSemicolonKV(s string) map[string]string {
	out := make(map[string]string)
	for part := range strings.SplitSeq(s, ";") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		k, v, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}
		out[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	if len(out) == 0 {
		return nil
	}
	return out
}

func clamp01(x float64) float64 {
	if x < 0 {
		return 0
	}
	if x > 1 {
		return 1
	}
	return x
}
