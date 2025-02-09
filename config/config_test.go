package config

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func TestNewFromFile(t *testing.T) {
	tests := []struct {
		name   string
		path   string
		expect *Config
		err    bool
	}{
		{
			name: "valid.yaml",
			path: "testdata/valid.yaml",
			expect: &Config{
				Interval: 60 * time.Second,
				Servers: []Server{
					{
						IP:   "50.108.13.235",
						Port: 2424,
					},
					{
						IP:   "50.108.13.235",
						Port: 2324,
					},
					{
						IP:   "50.108.13.235",
						Port: 2315,
					},
					{
						IP:   "50.108.13.235",
						Port: 27016,
					},
				},
			},
			err: false,
		},
		{
			name:   "invalid.yaml",
			path:   "testdata/invalid.yaml",
			expect: nil,
			err:    true,
		},
	}

	for _, tc := range tests {
		c, err := NewFromFile(tc.path)
		if tc.err {
			require.Error(t, err)
			return
		}
		require.NoError(t, err)
		require.Equal(t, tc.expect, c)
	}
}

func TestServerString(t *testing.T) {
	s := Server{
		IP:   "192.168.5.3",
		Port: 2304,
	}
	require.Equal(t, "192.168.5.3:2304", s.String())
}
