package json

import (
	"math/rand"
	"net/http"
	"testing"
	"time"

	"github.com/gekkotokio/golang-http-status-counter/counter"
)

func TestConvertRecordForJSONSource(t *testing.T) {
	codes := statusCodes()
	unixTime := time.Now().Unix()
	duration := 100
	ended := unixTime + int64(duration)
	looped := 10

	c := counter.Record{}

	rand.Seed(time.Now().UnixNano())

	for i := unixTime; i < ended; i++ {
		statuses := make(map[int]int)

		for j := 0; j < looped; j++ {
			idx := rand.Intn(len(codes))
			code := codes[idx]

			if _, ok := statuses[code]; !ok {
				statuses[code] = 100
			}

		}

		if len(statuses) == 0 {
			t.Errorf("expected length was over 0 but %v", len(statuses))
		}

		c[i] = statuses
	}

	if len(c) == 0 {
		t.Errorf("expected length of counter.Record was over 0 but %v", len(c))
	}

	sorted := sortRecordedTime(c)

	if len(sorted) == 0 {
		t.Errorf("expected length was %v but %v", duration, len(sorted))
	}

	source := convertRecordForJSONSource(c, sorted)

	if len(source) == 0 {
		t.Errorf("expected length was %v but %v", duration, len(source))
	}

	for _, record := range source {
		if record.RecordedAt != unixTime {
			t.Errorf("expected recorded unix time was %v but %v", unixTime, record.RecordedAt)
		}

		for _, status := range record.StatusCodes {
			if _, ok := c[unixTime][status.StatusCode]; !ok {
				t.Errorf("expected key %v of %v exsisted but not found", status.StatusCode, unixTime)
			} else if status.Counter != c[unixTime][status.StatusCode] {
				t.Errorf("expected the values were %v of %v but %v", c[unixTime][status.StatusCode], unixTime, status.Counter)
			}
		}

		unixTime++
	}
}

func TestSortRecordedTime(t *testing.T) {
	r := counter.Record{}

	var unixtime int64 = 101
	var looped int64 = 100

	for i := unixtime; i < unixtime+looped; i++ {
		r[i] = map[int]int{100: 200}
	}

	if len(r) < 1 {
		t.Errorf("expected counter.Record has some values but %v", len(r))
	}

	sorted := sortRecordedTime(r)

	if len(sorted) < 1 {
		t.Errorf("expected sorted has some values but %v", len(sorted))
	}

	for _, recordedAt := range sorted {
		if recordedAt != unixtime {
			t.Errorf("expected the values were %v but sorted value was %v", unixtime, recordedAt)
		}

		unixtime++
	}
}

func statusCodes() []int {
	return []int{
		http.StatusContinue,
		http.StatusSwitchingProtocols,
		http.StatusProcessing,
		http.StatusEarlyHints,
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
		http.StatusMultipleChoices,
		http.StatusMovedPermanently,
		http.StatusFound,
		http.StatusSeeOther,
		http.StatusNotModified,
		http.StatusUseProxy,
		http.StatusTemporaryRedirect,
		http.StatusPermanentRedirect,
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
