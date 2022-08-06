package proxy

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
)

func TestNewReverseProxy(t *testing.T) {
	a := make(AdditionalHeaders)
	message := "upstream response"

	backend := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		for key, value := range a {
			if _, ok := r.Header[key]; !ok {
				t.Errorf("expected the key exisisted but empty: %v", key)
			} else if !contains(r.Header[key], value) {
				t.Errorf("expected r.Header[%v] has %v but %v", key, value, r.Header[key][0])
			}
		}

		w.Write([]byte(message))
	}))
	defer backend.Close()

	host := "127.0.0.1"
	port, err := strconv.Atoi((strings.Split(backend.URL, ":"))[2])

	if err != nil {
		t.Errorf("expected got port number but %v", (strings.Split(backend.URL, ":"))[2])
	}

	c, err := NewConfig(HTTP, host, port)

	a["X-Forwarded-Proto"] = HTTP
	a["X-Forwarded-Host"] = "127.0.0.1"
	a["X-Forwarded-Port"] = c.port

	if err != nil {
		t.Errorf("expected no errors but occurred: %v", err.Error())
	}

	rp := NewReverseProxy(c, a)

	frontend := httptest.NewServer(rp)
	defer frontend.Close()

	request, err := http.NewRequest(http.MethodGet, frontend.URL, nil)

	if err != nil {
		t.Errorf("expected no errors but %v", err.Error())
		log.Fatal(err)
	}

	response, err := new(http.Client).Do(request)

	if err != nil {
		t.Errorf("expected no errors but %v", err.Error())
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	if err != nil {
		t.Errorf("expected no errors but %v", err.Error())
		log.Fatal(err)
	}

	if string(body) != message {
		t.Errorf("expected got %v but %v", message, string(body))
	}
}

func contains(values []string, target string) bool {
	for _, value := range values {
		if value == target {
			return true
		}
	}

	return false
}
