package consul

import (
	"net/http"
)

func NewHttpClient(d *Discovery) *http.Client {
	client := *http.DefaultClient
	client.Transport = &http.Transport{
		Proxy:       http.ProxyFromEnvironment,
		DialContext: DefaultDialer(d).DialContext,
	}
	return &client
}
