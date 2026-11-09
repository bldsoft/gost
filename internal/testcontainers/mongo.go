package testcontainers

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"os"
	"strings"

	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/mongodb"
	"github.com/testcontainers/testcontainers-go/wait"
)

const (
	mongoImage     = "mongo:8.2"
	replicaSetName = "rs0"
)

func StartMongo(ctx context.Context) (string, func(), error) {
	opts := []testcontainers.ContainerCustomizer{
		mongodb.WithReplicaSet(replicaSetName),
	}
	if tunables := os.Getenv("GLIBC_TUNABLES"); tunables != "" {
		opts = append(opts, testcontainers.WithEnv(map[string]string{
			"GLIBC_TUNABLES": tunables,
		}))
	}

	container, err := mongodb.Run(ctx, mongoImage, opts...)
	if err != nil {
		if container != nil {
			terminate(container)
		}

		return "", noop, fmt.Errorf("start mongodb container: %w", err)
	}

	if waitErr := wait.ForExec([]string{
		"mongosh", "--quiet", "--eval", "db.hello().isWritablePrimary",
	}).WithResponseMatcher(primaryReady).WaitUntilReady(ctx, container); waitErr != nil {
		terminate(container)

		return "", noop, fmt.Errorf("wait for mongodb primary: %w", waitErr)
	}

	rawURI, err := container.ConnectionString(ctx)
	if err != nil {
		terminate(container)

		return "", noop, fmt.Errorf("mongodb connection string: %w", err)
	}

	uri, err := withDirectConnection(rawURI)
	if err != nil {
		terminate(container)

		return "", noop, err
	}

	return uri, func() { terminate(container) }, nil
}

// The replica set advertises the container IP, which the host cannot dial.
func withDirectConnection(raw string) (string, error) {
	u, err := url.Parse(raw)
	if err != nil {
		return "", fmt.Errorf("parse mongodb uri: %w", err)
	}

	query := u.Query()
	query.Set("directConnection", "true")
	u.RawQuery = query.Encode()

	return u.String(), nil
}

func primaryReady(r io.Reader) bool {
	body, err := io.ReadAll(r)
	if err != nil {
		return false
	}

	return strings.TrimSpace(string(body)) == "true"
}
