package proxy

import (
	"net/http"
	"net/http/httputil"

	"github.com/gekkotokio/golang-http-status-counter/counter"
)

// NewReverseProxy returns httputil.ReverseProxy initialized by Config and AdditionalHeaders structs.
func NewReverseProxy(c Config, a AdditionalHeaders, m *counter.Measurement) *httputil.ReverseProxy {
	director := func(r *http.Request) {
		r.URL.Scheme = c.scheme

		if c.host == "" {
			r.URL.Host = c.host + c.port
		} else {
			r.URL.Host = c.port
		}

		if 0 < len(a) {
			for header, value := range a {
				r.Header.Set(header, value)
			}
		}
	}

	modififer := func(r *http.Response) error {
		m.CountUp(r.StatusCode)
		return nil
	}

	return &httputil.ReverseProxy{
		Director:       director,
		ModifyResponse: modififer,
	}
}
