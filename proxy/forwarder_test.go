package proxy

import (
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gekkotokio/golang-http-status-counter/counter"
)

func TestNewReverseProxy(t *testing.T) {
	a := make(AdditionalHeaders)
	message := "upstream response"

	// Generate a temporary test web server
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Test additonal HTTP headers from reverse proxy.
		for key, value := range a {
			if _, ok := r.Header[key]; !ok {
				t.Errorf("expected the key exisisted but empty: %v", key)
			} else if !contains(r.Header[key], value) {
				t.Errorf("expected r.Header[%v] has %v but %v", key, value, r.Header[key][0])
			}
		}

		w.Write([]byte(message))
	}))
	defer upstream.Close()

	// Parse information that the test web server is running
	host := "127.0.0.1"
	port, err := strconv.Atoi((strings.Split(upstream.URL, ":"))[2])

	if err != nil {
		t.Errorf("expected got port number but %v", (strings.Split(upstream.URL, ":"))[2])
	}

	// Prepare to generate the reverse proxy
	c, err := NewConfig(HTTP, host, port)

	a["X-Forwarded-Proto"] = HTTP
	a["X-Forwarded-Host"] = "127.0.0.1"
	a["X-Forwarded-Port"] = c.port

	if err != nil {
		t.Errorf("expected no errors but occurred: %v", err.Error())
	}

	// Initialize the reverse proxy
	rp := NewReverseProxy(c, a)

	// Run the test reverse proxy
	downstream := httptest.NewServer(rp)
	defer downstream.Close()

	request, err := http.NewRequest(http.MethodGet, downstream.URL, nil)

	if err != nil {
		t.Errorf("expected no errors but %v", err.Error())
		log.Fatal(err)
	}

	// Start tests
	response, err := new(http.Client).Do(request)

	if err != nil {
		t.Errorf("expected no errors but %v", err.Error())
		log.Fatal(err)
	}

	defer response.Body.Close()

	body, err := ioutil.ReadAll(response.Body)

	// Test HTTP response body from the upstream
	if err != nil {
		t.Errorf("expected no errors but %v", err.Error())
		log.Fatal(err)
	}

	if string(body) != message {
		t.Errorf("expected got %v but %v", message, string(body))
	}
}

// TestCountUpHTTPStatusCodes is the test that the number of HTTP status codes returned from the upstream matches the number of them counted by the reverse proxy
func TestCountUpHTTPStatusCodes(t *testing.T) {
	codes := statusCodes()
	expected := make(map[int]int)

	// Generate a temporary test web server
	// It returns randomed HTTP status code
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		code := 0

		rand.Seed(time.Now().UnixNano())

		for i := 0; i < 1; i++ {
			idx := rand.Intn(len(codes))
			code = codes[idx]
		}

		if _, ok := expected[code]; !ok {
			expected[code] = 0
		}

		expected[code]++

		w.WriteHeader(code)
	}))
	defer upstream.Close()

	// Prepare to generate the reverse proxy
	port, err := strconv.Atoi((strings.Split(upstream.URL, ":"))[2])

	if err != nil {
		t.Errorf("expected got port number but %v", (strings.Split(upstream.URL, ":"))[2])
	}

	c, err := NewConfig(HTTP, "", port)
	if err != nil {
		t.Errorf("expected no errors but occurred: %v", err.Error())
	}

	a := make(AdditionalHeaders)

	// Initialize the reverse proxy
	rp := NewReverseProxy(c, a)
	m := counter.NewMeasurement()

	// This map is a counter how many status codes are recieved in the reverse proxy
	called := make(map[int]int)

	mod := func(r *http.Response) error {
		if _, ok := called[r.StatusCode]; !ok {
			called[r.StatusCode] = 0
		}

		called[r.StatusCode]++

		m.CountUp(r.StatusCode)
		return nil
	}

	rp.ModifyResponse = mod

	// Run the test reverse proxy
	downstream := httptest.NewServer(rp)
	defer downstream.Close()

	max := 100
	// The map counts how many status codes are recieved by HTTP client
	actual := make(map[int]int)
	// The map aggregates HTTP status codes received by the reverse proxy
	counted := make(map[int]int)

	now := time.Now().Unix()

	for i := 0; i < max; i++ {
		request, err := http.NewRequest(http.MethodGet, downstream.URL, nil)

		if err != nil {
			t.Errorf("expected no errors but %v", err.Error())
			log.Fatal(err)
		}

		response, err := new(http.Client).Do(request)

		if err != nil {
			t.Errorf("expected no errors but %v", err.Error())
			log.Fatal(err)
		} else {
			if _, ok := actual[response.StatusCode]; !ok {
				actual[response.StatusCode] = 0
			}

			actual[response.StatusCode]++
		}
	}

	extracted, err := m.ExtractWithLockContext(now, now+1)

	if err != nil {
		t.Errorf("expected no errors but %v", err.Error())
	} else if len(extracted) == 0 {
		t.Errorf("expected length was over than 1 but %v", len(extracted))
	}

	for _, records := range extracted {
		for code, counter := range records {
			if _, ok := counted[code]; !ok {
				counted[code] = counter
			} else {
				counted[code] += counter
			}
		}
	}

	if len(expected) != len(actual) {
		t.Errorf("expected length was %v but %v", len(expected), len(actual))
	} else if len(counted) < len(expected) {
		// the length of counted should be more than the expected one
		t.Errorf("expected length was %v but %v", len(expected), len(counted))
		t.Errorf("extracted was %v", extracted)
		test, _ := m.GetRecordsAt(now)
		t.Errorf("Measurement was %v", test)
	}

	for code, counter := range expected {
		if counter != actual[code] {
			t.Errorf("expected %v were counted %v but got %v", code, counter, actual[code])
		}

		if counter != called[code] {
			t.Errorf("expected %v were counted %v but got %v", code, counter, called[code])
		}

		if counter != counted[code] {
			t.Errorf("expected %v were counted %v but got %v", code, counter, counted[code])
		}
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

func statusCodes() []int {
	codes := []int{
		// http.StatusContinue,
		// http.StatusSwitchingProtocols,
		// http.StatusProcessing,
		// http.StatusEarlyHints,
		http.StatusOK,
		http.StatusCreated,
		http.StatusAccepted,
		http.StatusNonAuthoritativeInfo,
		http.StatusNoContent,
		http.StatusResetContent,
		http.StatusPartialContent,
		http.StatusMultiStatus,
		http.StatusAlreadyReported,
		http.StatusIMUsed,
		// http.StatusMultipleChoices,
		// http.StatusMovedPermanently,
		// http.StatusFound,
		// http.StatusSeeOther,
		// http.StatusNotModified,
		// http.StatusUseProxy,
		// http.StatusTemporaryRedirect,
		// http.StatusPermanentRedirect,
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusPaymentRequired,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusMethodNotAllowed,
		http.StatusNotAcceptable,
		http.StatusProxyAuthRequired,
		http.StatusRequestTimeout,
		http.StatusConflict,
		http.StatusGone,
		http.StatusLengthRequired,
		http.StatusPreconditionFailed,
		http.StatusRequestEntityTooLarge,
		http.StatusRequestURITooLong,
		http.StatusUnsupportedMediaType,
		http.StatusRequestedRangeNotSatisfiable,
		http.StatusExpectationFailed,
		http.StatusTeapot,
		http.StatusMisdirectedRequest,
		http.StatusUnprocessableEntity,
		http.StatusLocked,
		http.StatusFailedDependency,
		http.StatusTooEarly,
		http.StatusUpgradeRequired,
		http.StatusPreconditionRequired,
		http.StatusTooManyRequests,
		http.StatusRequestHeaderFieldsTooLarge,
		http.StatusUnavailableForLegalReasons,
		http.StatusInternalServerError,
		http.StatusNotImplemented,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		http.StatusHTTPVersionNotSupported,
		http.StatusVariantAlsoNegotiates,
		http.StatusInsufficientStorage,
		http.StatusLoopDetected,
		http.StatusNotExtended,
		http.StatusNetworkAuthenticationRequired,
	}

	return codes
}
