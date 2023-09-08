package discovery

import (
	"context"
	"fmt"
	"time"
)

type Balancer[T any] interface {
	Balance(serviceCluster string, addrs []T) []T
}

type Resolver struct {
	discovery Discovery
	balancer  Balancer[string]
	cache     resolverCache
}

func NewResolver(discovery Discovery) *Resolver {
	return &Resolver{discovery: discovery, cache: newResolverCache(5 * time.Minute), balancer: &RoundRobin[string]{}}
}

func (r *Resolver) LookupServices(ctx context.Context, serviceName string) ([]string, error) {
	addrs, err := r.lookupServices(ctx, serviceName)
	if err != nil {
		return nil, err
	}
	if r.balancer != nil {
		addrs = r.balancer.Balance(serviceName, addrs)
	}
	return addrs, nil
}

func (r *Resolver) lookupServices(ctx context.Context, serviceName string) ([]string, error) {
	if addrs, err := r.cache.lookupServices(ctx, serviceName); err == nil && len(addrs) > 0 {
		return addrs, err
	}

	serviceInfo, err := r.discovery.ServiceByName(ctx, serviceName)

	if err != nil {
		return nil, err
	}

	addrs := make([]string, 0, len(serviceInfo.Instances))
	for _, info := range serviceInfo.Instances {
		addrs = append(addrs, fmt.Sprintf("%s:%d", info.Address, info.Port))
	}

	r.cache.put(serviceName, addrs)
	return addrs, nil
}
