package consul

import (
	"context"
	"fmt"

	"github.com/hashicorp/consul/api"
)

type Resolver struct {
	consulClient *api.Client
}

func NewResolver(consulClient *api.Client) *Resolver {
	return &Resolver{consulClient: consulClient}
}

func NewResolverFromDiscovery(discovery *Discovery) *Resolver {
	return NewResolver(discovery.ApiClient())
}

func (r *Resolver) LookupServices(ctx context.Context, serviceCluster string) ([]string, error) {
	_, checkInfos, err := r.consulClient.Agent().AgentHealthServiceByNameOpts(serviceCluster,
		&api.QueryOptions{
			Near: "_agent",
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

	return addrs, nil
}
