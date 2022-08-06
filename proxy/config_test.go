package proxy

import (
	"strconv"
	"testing"
)

func TestNewConfig(t *testing.T) {
	goodValues := []struct {
		host   string
		port   int
		scheme string
	}{
		{
			host:   "192.168.0.1",
			port:   8080,
			scheme: "http",
		},
		{
			host:   "192.168.0.1",
			port:   8080,
			scheme: "https",
		},
		{
			host:   "127.0.0.1",
			port:   8080,
			scheme: "http",
		},
		{
			host:   "127.0.0.1",
			port:   8080,
			scheme: "https",
		},
		{
			host:   "127.0.0.1",
			port:   9090,
			scheme: "http",
		},
		{
			host:   "127.0.0.1",
			port:   9090,
			scheme: "https",
		},
	}

	for _, v := range goodValues {
		if c, err := NewConfig(v.scheme, v.host, v.port); err != nil {
			t.Errorf("expected returned nil but %v", err.Error())
		} else if v.host != c.GetHost() {
			t.Errorf("expected the struct had %v but it had %v", v.host, c.host)
		} else if ":"+strconv.Itoa(v.port) != c.GetPort() {
			t.Errorf("expected the struct had %v but it had %v", v.port, c.port)
		} else if v.scheme != c.GetScheme() {
			t.Errorf("expected the struct had %v but it had %v", v.scheme, c.scheme)
		}
	}

	badValues := []struct {
		host   string
		port   int
		scheme string
	}{
		{
			host:   "192.168.0.1",
			port:   -1,
			scheme: "http",
		},
		{
			host:   "192.168.0.1",
			port:   65536,
			scheme: "https",
		},
		{
			host:   "127.0.0.1",
			port:   8080,
			scheme: "httpp",
		},
		{
			host:   "127.0.0.1",
			port:   8080,
			scheme: "httss",
		},
	}

	for _, v := range badValues {
		if _, err := NewConfig(v.scheme, v.host, v.port); err == nil {
			t.Errorf("expected returned error but nil")
		}
	}
}
