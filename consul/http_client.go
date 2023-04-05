package consul

import (
	"net/http"
)

func NewHttpClient(d *Discovery, sticky ...bool) *http.Client {
	client := *http.DefaultClient
	client.Transport = newTransport(d, sticky...)
	return &client
}

func newTransport(d *Discovery, sticky ...bool) http.RoundTripper {
	if len(sticky) > 0 && sticky[0] {
		return &http.Transport{
			Proxy:       http.ProxyFromEnvironment,
			DialContext: DefaultDialer(d).DialContext,
		}
	}
	return NewTransport(http.DefaultTransport, d)
}
