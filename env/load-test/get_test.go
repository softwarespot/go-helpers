package loadtest

import (
	"os"
	"testing"

	"github.com/softwarespot/go-helpers/env"
	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_Get(t *testing.T) {
	tests := []struct {
		name     string
		key      string
		fallback string
		want     string
	}{
		{
			name:     "exists",
			key:      "TEST_1",
			fallback: "fallback 1",
			want:     "variable 1",
		},
		{
			name:     "doesn't exist",
			key:      "TEST_3",
			fallback: "fallback 3",
			want:     "fallback 3",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer envCleanup()

			err := env.Load(os.DirFS("."), ".env.valid")
			testhelpers.AssertNoError(t, err)

			got := env.Get(tt.key, tt.fallback)
			testhelpers.AssertEqual(t, got, tt.want)
		})
	}
}
