package json

import (
	gojson "encoding/json"
	"sort"

	"github.com/gekkotokio/golang-http-status-counter/counter"
)

type Record []Counter

type Counter struct {
	RecordedAt  int64 `json:"recordedAt"`
	StatusCodes []struct {
		StatusCode int `json:"statusCode"`
		Counter    int `json:"counter"`
	} `json:"statusCodes"`
}

func ConvertRecord(r counter.Record) (json []byte, err error) {
	sorted := sortRecordedTime(r)
	source := convertRecordForJSONSource(r, sorted)

	json, err = gojson.Marshal(source)

	if err != nil {
		return nil, err
	}

	return json, nil
}

func convertRecordForJSONSource(r counter.Record, sorted []int64) Record {
	target := Record{}

	for _, unixTime := range sorted {
		c := Counter{RecordedAt: unixTime}

		if statuses, ok := r[unixTime]; ok {
			for code, counter := range statuses {
				c.StatusCodes = append(c.StatusCodes, struct {
					StatusCode int "json:\"statusCode\""
					Counter    int "json:\"counter\""
				}{
					StatusCode: code,
					Counter:    counter,
				})
			}
		}

		target = append(target, c)
	}

	return target
}

func sortRecordedTime(r counter.Record) []int64 {
	unixTime := []int64{}

	for recordedAt := range r {
		unixTime = append(unixTime, recordedAt)
	}

	sort.Slice(unixTime, func(i, j int) bool {
		return unixTime[i] < unixTime[j]
	})

	return unixTime
}
