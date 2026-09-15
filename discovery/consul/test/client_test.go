//go:build integration_test

package test

import (
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/bldsoft/gost/config"
	"github.com/bldsoft/gost/discovery"
	"github.com/bldsoft/gost/discovery/consul"
	"github.com/bldsoft/gost/server"
)

func runTestService(t *testing.T, serviceName, serviceID string) (cancel func()) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(serviceID))
	}))

	var consulCfg consul.Config
	consulCfg.SetDefaults()

	host, port, err := net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	assert.NoError(t, err)
	_, err = strconv.Atoi(port)
	assert.NoError(t, err)

	serverCfg := server.Config{
		ServiceName:     serviceName,
		ServiceInstance: serviceID,
		ServiceAddress:  config.Address("http://" + net.JoinHostPort(host, port)),
	}

	d := consul.NewDiscovery(serverCfg, consulCfg)
	assert.NoError(t, d.Register())

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
	clientDiscovery := consul.NewDiscovery(server.Config{ServiceName: "test-client"}, consulCfg)

	for _, sticky := range []bool{false} {
		t.Run(fmt.Sprintf("sticky=%v", sticky), func(t *testing.T) {
			httpClient := discovery.NewHttpClient(clientDiscovery, sticky)
			getResponseBody := func() string {
				resp, err := httpClient.Get("http://test/any")
				assert.NoError(t, err)
				defer func() { _ = resp.Body.Close() }()
				data, err := io.ReadAll(resp.Body)
				assert.NoError(t, err)

				return string(data)
			}

			expected := "111111"
			if !sticky {
				expected = "121212"
			}

			actual := ""
			for range expected {
				actual += getResponseBody()
			}
			assert.Equal(t, expected, actual)
		})
	}
}
