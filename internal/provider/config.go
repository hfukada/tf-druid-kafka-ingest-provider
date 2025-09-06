package provider

import (
	"net/http"
	"time"
)

type Config struct {
	Endpoint string
	Username string
	Password string
	Timeout  int
}

type Client struct {
	HTTPClient *http.Client
	Endpoint   string
	Username   string
	Password   string
}

func (c *Config) Client() (*Client, error) {
	client := &Client{
		HTTPClient: &http.Client{
			Timeout: time.Duration(c.Timeout) * time.Second,
		},
		Endpoint: c.Endpoint,
		Username: c.Username,
		Password: c.Password,
	}

	return client, nil
}