//go:build integration_test

package test

import (
	"context"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/consul"
	"github.com/bldsoft/gost/internal/testcontainers"
	"github.com/bldsoft/gost/server"
)

var consulAddr config.HttpAddress

func TestMain(m *testing.M) {
	raw, cleanup, err := testcontainers.StartConsul(context.Background())
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "%v\n", err)
		os.Exit(1)
	}
	consulAddr = config.HttpAddress(raw)

	code := m.Run()
	cleanup()
	os.Exit(code)
}

func runTestService(t *testing.T, serviceName, serviceID string) (cancel func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(serviceID))
	}))

	var consulCfg consul.Config
	consulCfg.SetDefaults()
	consulCfg.ConsulAddr = consulAddr

	host, port, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	require.NoError(t, err)
	_, err = strconv.Atoi(port)
	require.NoError(t, err)

	serverCfg := server.Config{
		ServiceName:     serviceName,
		ServiceInstance: serviceID,
		ServiceAddress:  config.Address("http://" + net.JoinHostPort(host, port)),
	}

	d := consul.NewDiscovery(serverCfg, consulCfg)
	require.NoError(t, d.Register())

	return func() {
		srv.Close()
		_ = d.Deregister()
	}
}

func TestClient(t *testing.T) {
	defer runTestService(t, "test", "1")()
	defer runTestService(t, "test", "2")()

	var consulCfg consul.Config
	consulCfg.SetDefaults()
	consulCfg.ConsulAddr = consulAddr
	clientDiscovery := consul.NewDiscovery(server.Config{ServiceName: "test-client"}, consulCfg)

	for _, sticky := range []bool{false} {
		t.Run(fmt.Sprintf("sticky=%v", sticky), func(t *testing.T) {
			httpClient := discovery.NewHttpClient(clientDiscovery, sticky)
			getResponseBody := func() string {
				resp, err := httpClient.Get("http://test/any")
				require.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				data, err := io.ReadAll(resp.Body)
				require.NoError(t, err)

				return string(data)
			}

			actual := ""
			for range 6 {
				actual += getResponseBody()
			}

			// order not guaranteed; underlying map has randomized output.
			if sticky {
				assert.Contains(t, []string{"111111", "222222"}, actual)

				return
			}
			assert.Contains(t, []string{"121212", "212121"}, actual)
		})
	}
}
