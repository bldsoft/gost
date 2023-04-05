// // go:build integration_test

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

	"github.com/bldsoft/gost/consul"
	"github.com/hashicorp/consul/api"
	"github.com/stretchr/testify/assert"
)

func runTestService(t *testing.T, cluster, serviceID string) (cancel func()) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(serviceID))
	}))

	var cfg consul.Config
	cfg.ConsulConfig.SetDefaults()
	cfg.ServiceConfig.SetDefaults()
	cfg.ServiceID = serviceID
	cfg.Cluster = cluster
	var (
		err  error
		port string
	)
	cfg.ServiceAddr, port, err = net.SplitHostPort(strings.TrimPrefix(srv.URL, "http://"))
	assert.NoError(t, err)
	cfg.ServicePort, err = strconv.Atoi(port)
	assert.NoError(t, err)

	consul.Register(cfg)

	return func() {
		srv.Close()
	}
}

func TestClient(t *testing.T) {
	defer runTestService(t, "test", "1")()
	defer runTestService(t, "test", "2")()

	client, err := api.NewClient(api.DefaultConfig())
	assert.NoError(t, err)

	for _, sticky := range []bool{false} {
		t.Run(fmt.Sprintf("sticky=%v", sticky), func(t *testing.T) {
			httpClient := consul.NewHttpClient(client, sticky)
			getResponseBody := func() string {
				resp, err := httpClient.Get("http://test/any")
				assert.NoError(t, err)
				defer resp.Body.Close()
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
