// Package client provides a client for interacting
// with DZSA.
package client

import (
	"encoding/json"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/jsirianni/dayzsa-exporter/model"
)

const (
	baseURL = "https://dayzsalauncher.com/api/v1/query"

	clientRequestTimeout = 10 * time.Second
)

// New creates a new client.
func New() (Client, error) {
	c := &defaultClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: clientRequestTimeout,
		},
	}

	return c, nil
}

// Client is a client for interacting with DZSA.
type Client interface {
	Query(ip string, port int) (*model.QueryResponse, error)
}

type defaultClient struct {
	baseURL string
	client  *http.Client
}

var _ Client = &defaultClient{}

// Query queries the DZSA api and returns a query response
// It is up to the caller to ensure ip and port are valid.
func (c *defaultClient) Query(ip string, port int) (*model.QueryResponse, error) {
	endpoint, err := buildEndpoint(c.baseURL, ip, port)
	if err != nil {
		return nil, fmt.Errorf("build endpoint: %w", err)
	}

	req, err := http.NewRequest(http.MethodGet, endpoint, nil)
	if err != nil {
		return nil, fmt.Errorf("create request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("do request: %w", err)
	}
	defer resp.Body.Close()

	queryResponse := &model.QueryResponse{}
	if err := json.NewDecoder(resp.Body).Decode(&queryResponse); err != nil {
		return nil, fmt.Errorf("decode response: %w", err)
	}

	return queryResponse, nil
}

func buildEndpoint(base, ip string, port int) (string, error) {
	u, err := url.Parse(base)
	if err != nil {
		return "", fmt.Errorf("parse base: %w", err)
	}

	ipPort := net.JoinHostPort(ip, strconv.Itoa(port))

	path, err := url.JoinPath(u.Path, ipPort)
	if err != nil {
		return "", fmt.Errorf("join path: %s: %w", ipPort, err)
	}

	u.Path = path

	return u.String(), nil
}
