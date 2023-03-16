package breaker

import (
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	http.Client
	circuitBreaker *CircuitBreaker
}

func NewClient(c http.Client, settings settings) *Client {
	return &Client{Client: c, circuitBreaker: NewCircuitBreaker(settings)}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.Do(req)
	})
	return resp.(*http.Response), err
}

func (c *Client) Get(url string) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.Get(url)
	})
	return resp.(*http.Response), err
}

func (c *Client) Head(url string) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.Head(url)
	})
	return resp.(*http.Response), err
}

func (c *Client) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.Post(url, contentType, body)
	})
	return resp.(*http.Response), err
}

func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.PostForm(url, data)
	})
	return resp.(*http.Response), err
}
