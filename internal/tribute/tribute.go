package tribute

import (
	"net/http"
)

type Client struct{}

func NewClient(_ interface{}, _ interface{}) *Client { return &Client{} }

func (c *Client) WebHookHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})
}
