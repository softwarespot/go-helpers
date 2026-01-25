package loadtest

import (
	"os"
	"testing"

	"github.com/softwarespot/go-helpers/env"
	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_Load(t *testing.T) {
	tests := []struct {
		name    string
		envFile string
		want    map[string]string
		wantErr bool
	}{
		{
			name:    "missing .env file",
			envFile: ".env.missing",
			want:    nil,
			wantErr: true,
		},
		{
			name:    "invalid .env file",
			envFile: ".env.invalid",
			want: map[string]string{
				"TEST_1": "variable 1",
			},
			wantErr: true,
		},
		{
			name:    "load .env file",
			envFile: ".env.valid",
			want: map[string]string{
				"TEST_1": "variable 1",
				"TEST_2": "",
				"Test3":  "1234",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer envCleanup()

			err := env.Load(os.DirFS("."), tt.envFile)
			testhelpers.AssertEqual(t, (err != nil), tt.wantErr)

			for key, want := range tt.want {
				v, ok := os.LookupEnv(key)
				testhelpers.AssertEqual(t, v, want)
				testhelpers.AssertEqual(t, ok, want != "")
			}
		})
	}
}
