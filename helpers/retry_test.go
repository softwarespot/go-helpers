package helpers

import (
	"errors"
	"testing"
	"time"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_Retry(t *testing.T) {
	iterations := 0
	err := errors.New("unexpected error")
	tests := []struct {
		name             string
		fn               func(iter int) error
		retries          int
		wantIterCombined int
		wantErr          bool
	}{
		{
			name: "function should be called once when retries is 0",
			fn: func(iter int) error {
				iterations += iter
				return err
			},
			retries:          0,
			wantIterCombined: 1,
			wantErr:          true,
		},
		{
			name: "function should be called once when retries less than 0",
			fn: func(iter int) error {
				iterations += iter
				return err
			},
			retries:          -2,
			wantIterCombined: 1,
			wantErr:          true,
		},
		{
			name: "function should be called maximum number of retries",
			fn: func(iter int) error {
				iterations += iter
				return err
			},
			retries:          3,
			wantIterCombined: 6,
			wantErr:          true,
		},
		{
			name: "function should be called once when no error occurs",
			fn: func(iter int) error {
				iterations += iter
				return nil
			},
			retries:          3,
			wantIterCombined: 1,
			wantErr:          false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			iterations = 0
			err := Retry(tt.fn, tt.retries, 1*time.Microsecond)
			testhelpers.AssertEqual(t, err != nil, tt.wantErr)
			testhelpers.AssertEqual(t, iterations, tt.wantIterCombined)
		})
	}
}
