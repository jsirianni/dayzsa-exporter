package model

import "testing"

func TestEndpointString(t *testing.T) {
	e := Endpoint{
		IP:   "10.99.1.10",
		Port: 2302,
	}

	want := "10.99.1.10:2302"
	got := e.String()
	if got != want {
		t.Errorf("want %s, got %s", want, got)
	}
}
