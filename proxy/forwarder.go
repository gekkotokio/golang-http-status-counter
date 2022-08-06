package proxy

import (
	"net/http"
	"net/http/httputil"
)

// NewReverseProxy returns httputil.ReverseProxy initialized by Config and AdditionalHeaders structs.
func NewReverseProxy(c Config, a AdditionalHeaders) *httputil.ReverseProxy {
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

	return &httputil.ReverseProxy{Director: director}
}
