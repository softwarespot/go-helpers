package backoff

import (
	"testing"
	"time"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_Exponential(t *testing.T) {
	maxAttempt := 0
	var ds []time.Duration
	for attempt, d := range Exponential(WithRetryLimit(5)) {
		maxAttempt = attempt
		ds = append(ds, d)
	}

	testhelpers.AssertEqual(t, ds, []time.Duration{
		1 * time.Second,
		2 * time.Second,
		4 * time.Second,
		8 * time.Second,
		16 * time.Second,
	})
	testhelpers.AssertEqual(t, maxAttempt, 5)

	for attempt, d := range Exponential(WithRetryLimit(5), WithJitter()) {
		t.Log(attempt, d)
	}
}
