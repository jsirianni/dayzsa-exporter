// Package ifconfig is used to detect the public
// ip address of the host
package ifconfig

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"go.uber.org/zap"
)

const (
	endpoint = "https://ifconfig.net/json"

	clientRequestTimeout = 30 * time.Second
)

// Response is the response from the ifconfig.net service
type Response struct {
	IP         string  `json:"ip"`
	IPDecimal  int     `json:"ip_decimal"`
	Country    string  `json:"country"`
	CountryIso string  `json:"country_iso"`
	CountryEu  bool    `json:"country_eu"`
	Latitude   float64 `json:"latitude"`
	Longitude  float64 `json:"longitude"`
	TimeZone   string  `json:"time_zone"`
	Asn        string  `json:"asn"`
	AsnOrg     string  `json:"asn_org"`
	Hostname   string  `json:"hostname"`
	UserAgent  struct {
		Product  string `json:"product"`
		Version  string `json:"version"`
		Comment  string `json:"comment"`
		RawValue string `json:"raw_value"`
	} `json:"user_agent"`
}

// New takes a HTTP client creates a new ifconfig client
func New(logger *zap.Logger) Client {
	return Client{
		client: &http.Client{
			Timeout: clientRequestTimeout,
		},
		logger: logger,
	}
}

// Client is the ifconfig client
type Client struct {
	client  *http.Client
	logger  *zap.Logger
	address string
	mu      sync.Mutex
}

// Get gets the public ip address of the host
func (c *Client) Get() (*Response, error) {
	resp := Response{}
	req, err := http.NewRequest("GET", endpoint, nil)
	if err != nil {
		return nil, err
	}

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer res.Body.Close()

	if res.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", res.StatusCode)
	}

	if err := json.NewDecoder(res.Body).Decode(&resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// Start starts the ifconfig client loop
func (c *Client) Start(ctx context.Context) error {
	ticker := time.NewTicker(60 * time.Second)
	defer ticker.Stop()

	resp, err := c.Get()
	if err != nil {
		return fmt.Errorf("startup get: %w", err)
	}

	if resp.IP == "" {
		return fmt.Errorf("empty ip address")
	}

	c.mu.Lock()
	c.address = resp.IP
	c.mu.Unlock()

	c.logger.Info("public ip address updated", zap.String("ip", resp.IP))

	c.logger.Info("starting request loop")
	go func() {
		for {
			select {
			case <-ticker.C:
				resp, err := c.Get()
				if err != nil {
					c.logger.Error("get", zap.Error(err))
					continue
				}

				if resp.IP == "" {
					c.logger.Error("empty ip address")
					continue
				}

				c.mu.Lock()
				c.address = resp.IP
				c.mu.Unlock()
				c.logger.Debug("public ip address updated", zap.String("ip", resp.IP))

			case <-ctx.Done():
				c.logger.Info("shutting down")
				return
			}
		}
	}()

	return nil
}

// GetAddress gets the public ip address of the host
func (c *Client) GetAddress() string {
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.address
}
