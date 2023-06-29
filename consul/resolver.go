package consul

import (
	"context"
	"fmt"
	"time"

	"github.com/hashicorp/consul/api"
)

type Balancer[T any] interface {
	Balance(serviceCluster string, addrs []T) []T
}

type Resolver struct {
	consulClient *api.Client
	balancer     Balancer[string]
	cache        resolverCache
}

func NewResolver(consulClient *api.Client) *Resolver {
	return &Resolver{consulClient: consulClient, cache: newResolverCache(5 * time.Minute), balancer: &RoundRobin[string]{}}
}

func NewResolverFromDiscovery(discovery *Discovery) *Resolver {
	return NewResolver(discovery.ApiClient())
}

func (r *Resolver) LookupServices(ctx context.Context, serviceCluster string) ([]string, error) {
	addrs, err := r.lookupServices(ctx, serviceCluster)
	if err != nil {
		return nil, err
	}
	if r.balancer != nil {
		addrs = r.balancer.Balance(serviceCluster, addrs)
	}
	return addrs, nil
}

func (r *Resolver) lookupServices(ctx context.Context, serviceCluster string) ([]string, error) {
	if addrs, err := r.cache.lookupServices(ctx, serviceCluster); err == nil && len(addrs) > 0 {
		return addrs, err
	}

	_, checkInfos, err := r.consulClient.Agent().AgentHealthServiceByNameOpts(serviceCluster,
		&api.QueryOptions{
			// Near: "_agent",
			// TODO: filter status "passing"
		},
	)

	if err != nil {
		return nil, err
	}

	addrs := make([]string, 0, len(checkInfos))

	for _, addr := range checkInfos {
		addrs = append(addrs, fmt.Sprintf("%s:%d", addr.Service.Address, addr.Service.Port))
	}

	r.cache.put(serviceCluster, addrs)

	return addrs, nil
}
