package breaker

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
)

type Client struct {
	*http.Client
	circuitBreaker *CircuitBreaker
}

func NewClient(c *http.Client, settings settings) *Client {
	if settings.isSuccessful == nil {
		settings = settings.WithIsSuccessful(func(result any, err error) error {
			if err != nil {
				return err
			}
			resp := result.(*http.Response)
			if resp.StatusCode >= 500 {
				return fmt.Errorf("%d %s", resp.StatusCode, http.StatusText(resp.StatusCode))
			}
			return nil
		})
	}
	return &Client{Client: c, circuitBreaker: NewCircuitBreaker(settings)}
}

func (c *Client) Do(req *http.Request) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.Do(req)
	})
	if err != nil {
		return nil, err
	}
	return resp.(*http.Response), err
}

func (c *Client) Get(url string) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.Get(url)
	})
	if err != nil {
		return nil, err
	}
	return resp.(*http.Response), err
}

func (c *Client) Head(url string) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.Head(url)
	})
	if err != nil {
		return nil, err
	}
	return resp.(*http.Response), err
}

func (c *Client) Post(url string, contentType string, body io.Reader) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.Post(url, contentType, body)
	})
	if err != nil {
		return nil, err
	}
	return resp.(*http.Response), err
}

func (c *Client) PostForm(url string, data url.Values) (*http.Response, error) {
	resp, err := c.circuitBreaker.Execute(func() (interface{}, error) {
		return c.Client.PostForm(url, data)
	})
	if err != nil {
		return nil, err
	}
	return resp.(*http.Response), err
}
