package consul

import (
	"net/http"

	"github.com/hashicorp/consul/api"
)

func NewHttpClient(consulClient *api.Client, sticky ...bool) *http.Client {
	client := *http.DefaultClient
	client.Transport = newTransport(consulClient, sticky...)
	return &client
}

func NewHttpClientFromDiscovery(d *Discovery, sticky ...bool) *http.Client {
	return NewHttpClient(d.ApiClient(), sticky...)
}

func newTransport(consulClient *api.Client, sticky ...bool) http.RoundTripper {
	if len(sticky) > 0 && sticky[0] {
		return &http.Transport{
			Proxy:       http.ProxyFromEnvironment,
			DialContext: DefaultDialer(consulClient).DialContext,
		}
	}
	return NewTransport(http.DefaultTransport, consulClient)
}
