package proxy

import (
	"fmt"
	"strconv"
)

const (
	HTTP          string = "http"
	HTTPS         string = "https"
	MaxPortNumber int    = 65535
	MinPortNumber int    = 0
)

// Config is the struct to set the values taht calls httputil.ReverseProxy.Director
type Config struct {
	host   string
	port   string
	scheme string
}

// NewConfig returns initialized Config struct.
// It contains the information to forward HTTP request to the upstream.
func NewConfig(scheme string, host string, port int) (Config, error) {
	if err := setScheme(scheme); err != nil {
		return Config{}, err
	}

	if err := setForwardPort(port); err != nil {
		return Config{}, err
	}

	c := Config{
		host:   host,
		port:   ":" + strconv.Itoa(port),
		scheme: scheme,
	}

	return c, nil
}

// GetHost returns hostname or IP adoress to forward HTTP request to the upstream.
func (c *Config) GetHost() string {
	return c.host
}

// GetPort returns the port number that the upstream is listening on.
func (c *Config) GetPort() string {
	return c.port
}

// GetScheme returns the schema that connects to the upstream.
func (c *Config) GetScheme() string {
	return c.scheme
}

func setScheme(scheme string) error {
	if scheme == HTTP || scheme == HTTPS {
		return nil
	}

	return fmt.Errorf("the given value was invalid scheme: %v", scheme)
}

func setForwardPort(port int) error {
	if port < MinPortNumber || MaxPortNumber < port {
		return fmt.Errorf("the given value %v was out of range, it should be between 0 and 65535", port)
	}

	return nil
}
