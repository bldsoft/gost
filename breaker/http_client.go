package breaker

import (
	"net/http"
)

type Client struct {
	http.Client
	circuitBreaker *CircuitBreaker
}

// func NewClient(c http.Client) *Client {
// 	return &Client{Client: c, circuitBreaker: NewCircuitBreaker()}
// }

// func (c *Client) Do(req *http.Request) (*http.Response, error) {

// }

// func (c *Client) Get(url string) (resp *http.Response, err error) {

// }

// func (c *Client) Head(url string) (resp *http.Response, err error) {

// }

// func (c *Client) Post(url string, contentType string, body io.Reader) (resp *http.Response, err error) {

// }

// func (c *Client) PostForm(url string, data url.Values) (resp *http.Response, err error) {

// }
