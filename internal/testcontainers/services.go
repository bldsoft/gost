package testcontainers

import (
	"context"
	"fmt"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	memcachedImage = "memcached:1.6.5"
	memcachedPort  = "11211/tcp"

	consulImage = "hashicorp/consul:1.15.4"
	consulPort  = "8500/tcp"
)

func StartMemcached(ctx context.Context) (string, func(), error) {
	container, err := testcontainers.Run(ctx, memcachedImage,
		testcontainers.WithExposedPorts(memcachedPort),
		testcontainers.WithWaitStrategy(wait.ForListeningPort(memcachedPort)),
	)
	if err != nil {
		if container != nil {
			terminate(container)
		}

		return "", noop, fmt.Errorf("start memcached container: %w", err)
	}

	addr, err := container.PortEndpoint(ctx, memcachedPort, "")
	if err != nil {
		terminate(container)

		return "", noop, fmt.Errorf("memcached address: %w", err)
	}

	return addr, func() { terminate(container) }, nil
}

func StartConsul(ctx context.Context) (string, func(), error) {
	container, err := testcontainers.Run(ctx, consulImage,
		testcontainers.WithExposedPorts(consulPort),
		testcontainers.WithCmd("agent", "-dev", "-client=0.0.0.0"),
		testcontainers.WithWaitStrategy(wait.ForHTTP("/v1/status/leader").WithPort(consulPort)),
	)
	if err != nil {
		if container != nil {
			terminate(container)
		}

		return "", noop, fmt.Errorf("start consul container: %w", err)
	}

	addr, err := container.PortEndpoint(ctx, consulPort, "http")
	if err != nil {
		terminate(container)

		return "", noop, fmt.Errorf("consul address: %w", err)
	}

	return addr, func() { terminate(container) }, nil
}
