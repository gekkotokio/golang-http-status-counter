package proxy

import (
	"bytes"
	cryptorand "crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"io/ioutil"
	"log"
	"math/big"
	"math/rand"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"sync"
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
	m := counter.NewMeasurement()
	rp := NewReverseProxy(c, a, &m)

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
	var mu sync.Mutex

	// Generate a temporary test web server
	// It returns randomed HTTP status code
	upstream := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			mu.Unlock()
		}()

		code := 0

		rand.Seed(time.Now().UnixNano())

		idx := rand.Intn(len(codes))
		code = codes[idx]

		time.Sleep(100 * time.Microsecond)

		mu.Lock()

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
	m := counter.NewMeasurement()
	rp := NewReverseProxy(c, a, &m)

	// Run the test reverse proxy
	downstream := httptest.NewServer(rp)
	defer downstream.Close()

	// The map counts how many status codes are recieved by HTTP client
	actual := make(map[int]int)
	// The map aggregates HTTP status codes received by the reverse proxy
	counted := make(map[int]int)

	now := time.Now().Unix()
	var wg sync.WaitGroup
	duration := 5
	loop := 100

	for i := 0; i < duration; i++ {
		for j := 0; j < loop; j++ {
			wg.Add(1)

			go func(wg *sync.WaitGroup) {
				defer func() {
					wg.Done()
				}()
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
					mu.Lock()
					if _, ok := actual[response.StatusCode]; !ok {
						actual[response.StatusCode] = 0
					}

					actual[response.StatusCode]++
					mu.Unlock()
				}
			}(&wg)
		}

		time.Sleep(1 * time.Second)
	}

	wg.Wait()

	extracted, err := m.ExtractWithLockContext(now, now+int64(duration))

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
		test, _ := m.GetRecordsAtWithLockContext(now)
		t.Errorf("Measurement was %v", test)
	}

	for code, counter := range expected {
		if counter != actual[code] {
			t.Errorf("expected %v were counted %v but got %v", code, counter, actual[code])
		}

		if counter != counted[code] {
			t.Errorf("expected %v were counted %v but got %v", code, counter, counted[code])
		}
	}
}

func TestConnectWithHTTPS(t *testing.T) {
	// Get test CA and server certificate
	serverTLSConf, clientTLSConf, err := certsetup()

	if err != nil {
		t.Errorf("expected no errors but %v", err.Error())
		panic(err)
	}

	// Generate a temporary test web server
	// It returns randomed HTTP status code

	codes := statusCodes()
	expected := make(map[int]int)
	var mu sync.Mutex

	upstream := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			mu.Unlock()
		}()

		code := 0

		rand.Seed(time.Now().UnixNano())

		idx := rand.Intn(len(codes))
		code = codes[idx]

		time.Sleep(100 * time.Microsecond)

		mu.Lock()

		if _, ok := expected[code]; !ok {
			expected[code] = 0
		}

		expected[code]++

		w.WriteHeader(code)
	}))

	upstream.TLS = serverTLSConf
	upstream.StartTLS()
	defer upstream.Close()

	// Prepare to generate the reverse proxy
	port, err := strconv.Atoi((strings.Split(upstream.URL, ":"))[2])

	if err != nil {
		t.Errorf("expected got port number but %v", (strings.Split(upstream.URL, ":"))[2])
	}

	c, err := NewConfig(HTTPS, "", port)
	if err != nil {
		t.Errorf("expected no errors but occurred: %v", err.Error())
	}

	a := make(AdditionalHeaders)

	// Initialize the reverse proxy
	m := counter.NewMeasurement()
	rp := NewReverseProxy(c, a, &m)

	transport := &http.Transport{
		TLSClientConfig: clientTLSConf,
	}

	// rp.Transport = transport

	rp.Transport = &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}

	// Run the test reverse proxy
	downstream := httptest.NewUnstartedServer(rp)
	downstream.TLS = serverTLSConf
	downstream.StartTLS()
	defer downstream.Close()

	// The map counts how many status codes are recieved by HTTP client
	actual := make(map[int]int)
	// The map aggregates HTTP status codes received by the reverse proxy
	counted := make(map[int]int)

	now := time.Now().Unix()
	var wg sync.WaitGroup
	duration := 5
	loop := 100

	// Initialize the test HTTP client
	client := downstream.Client()
	client.Transport = transport

	for i := 0; i < duration; i++ {
		for j := 0; j < loop; j++ {
			wg.Add(1)

			go func(wg *sync.WaitGroup) {
				defer func() {
					wg.Done()
				}()

				// request, err := http.NewRequest(http.MethodGet, downstream.URL, nil)
				response, err := client.Get(downstream.URL)

				if err != nil {
					t.Errorf("expected no errors but %v", err.Error())
					log.Fatal(err)
				} else {
					mu.Lock()
					if _, ok := actual[response.StatusCode]; !ok {
						actual[response.StatusCode] = 0
					}

					actual[response.StatusCode]++
					mu.Unlock()
				}
			}(&wg)
		}

		time.Sleep(1 * time.Second)
	}

	wg.Wait()

	extracted, err := m.ExtractWithLockContext(now, now+int64(duration))

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
		test, _ := m.GetRecordsAtWithLockContext(now)
		t.Errorf("Measurement was %v", test)
	}

	for code, counter := range expected {
		if counter != actual[code] {
			t.Errorf("expected %v were counted %v but got %v", code, counter, actual[code])
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
	return []int{
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
}

func certsetup() (serverTLSConf *tls.Config, clientTLSConf *tls.Config, err error) {
	yesterday := time.Now().Add(-24 * time.Hour)
	tomorrow := time.Now().Add(24 * time.Hour)

	// Set up test CA certificate
	// https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.9
	ca := &x509.Certificate{
		BasicConstraintsValid: true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		IsCA:                  true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		// https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.9
		NotAfter:     time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 99, time.UTC),
		NotBefore:    time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC),
		SerialNumber: big.NewInt(int64(yesterday.Year())),
		Subject: pkix.Name{
			CommonName:         "John Doe",
			Country:            []string{"JP"},
			Organization:       []string{"Organization for test"},
			OrganizationalUnit: []string{"Organization Unit for test"},
			PostalCode:         []string{"123-4567"},
			Province:           []string{"Tokyo"},
			StreetAddress:      []string{"Unknown City 890-123"},
		},
	}

	// Create test private and public key
	privateCAKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)

	if err != nil {
		return nil, nil, err
	}

	// Create the CA
	caBytes, err := x509.CreateCertificate(cryptorand.Reader, ca, ca, &privateCAKey.PublicKey, privateCAKey)

	if err != nil {
		return nil, nil, err
	}

	// PEM encode
	caPEM := new(bytes.Buffer)

	pem.Encode(caPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})

	caPrivKeyPEM := new(bytes.Buffer)

	pem.Encode(caPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateCAKey),
	})

	// Set up test server certificate
	cert := &x509.Certificate{
		//DNSNames:    []string{"localhost"},
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		IPAddresses: []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback},
		KeyUsage:    x509.KeyUsageDigitalSignature,
		// https://datatracker.ietf.org/doc/html/rfc5280#section-4.2.1.9
		NotAfter:     time.Date(tomorrow.Year(), tomorrow.Month(), tomorrow.Day(), 23, 59, 59, 99, time.UTC),
		NotBefore:    time.Date(yesterday.Year(), yesterday.Month(), yesterday.Day(), 0, 0, 0, 0, time.UTC),
		SerialNumber: big.NewInt(int64(yesterday.Year())),
		Subject: pkix.Name{
			CommonName:         "John Doe",
			Country:            []string{"JP"},
			Organization:       []string{"Organization for test"},
			OrganizationalUnit: []string{"Organization Unit for test"},
			PostalCode:         []string{"123-4567"},
			Province:           []string{"Tokyo"},
			StreetAddress:      []string{"Unknown City 890-123"},
		},
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
	}

	certPrivacyKey, err := rsa.GenerateKey(cryptorand.Reader, 4096)
	if err != nil {
		return nil, nil, err
	}

	certBytes, err := x509.CreateCertificate(cryptorand.Reader, cert, ca, &certPrivacyKey.PublicKey, privateCAKey)

	if err != nil {
		return nil, nil, err
	}

	certPEM := new(bytes.Buffer)

	pem.Encode(certPEM, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: certBytes,
	})

	certPrivKeyPEM := new(bytes.Buffer)

	pem.Encode(certPrivKeyPEM, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(certPrivacyKey),
	})

	serverCert, err := tls.X509KeyPair(certPEM.Bytes(), certPrivKeyPEM.Bytes())

	if err != nil {
		return nil, nil, err
	}

	serverTLSConf = &tls.Config{
		Certificates: []tls.Certificate{serverCert},
	}

	certpool := x509.NewCertPool()

	certpool.AppendCertsFromPEM(caPEM.Bytes())

	clientTLSConf = &tls.Config{
		RootCAs: certpool,
	}

	return
}
