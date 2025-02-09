package client

import (
	"testing"
)

func Test_buildEndpoint(t *testing.T) {
	tests := []struct {
		name    string
		baseURL string
		ip      string
		port    int
		want    string
	}{
		{
			name:    "valid",
			baseURL: "https://dayzsalauncher.com/api/v1/query",
			ip:      "50.108.13.235",
			port:    2424,
			want:    "https://dayzsalauncher.com/api/v1/query/50.108.13.235:2424",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			got, err := buildEndpoint(tc.baseURL, tc.ip, tc.port)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tc.want {
				t.Errorf("got: %s, want: %s", got, tc.want)
			}
		})
	}
}
