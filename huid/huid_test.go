package huid

import (
	"testing"
	"time"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_New(t *testing.T) {
	originalNow := nowFn
	t.Cleanup(func() {
		nowFn = originalNow
	})
	SetNowFunc(func() time.Time {
		return time.Date(2025, 12, 0o6, 10, 1, 14, 1000, time.Local)
	})
	testhelpers.AssertEqual(t, New(), "20251206-100114")
}
