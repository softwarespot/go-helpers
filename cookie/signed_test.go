package cookie

import (
	"net/http/httptest"
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_NewSigned(t *testing.T) {
	tests := []struct {
		name      string
		secret    string
		want      []byte
		wantPanic bool
	}{
		{
			name:      "valid SHA-256 secret",
			secret:    "4366d3f57f71049774c039609100ea220467062dfa6eeed93a939629c173ad5d",
			want:      []byte{67, 102, 211, 245, 127, 113, 4, 151, 116, 192, 57, 96, 145, 0, 234, 34, 4, 103, 6, 45, 250, 110, 238, 217, 58, 147, 150, 41, 193, 115, 173, 93},
			wantPanic: false,
		},
		{
			name:      "invalid secret Length",
			secret:    "a3c2f4e5d6b7",
			wantPanic: true,
		},
		{
			name:      "invalid hexadecimal Secret",
			secret:    "invalidHexSecret",
			wantPanic: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.wantPanic {
				defer func() {
					err := recover()
					testhelpers.AssertError(t, err.(error))
				}()
			}

			signed := NewSigned(tt.secret, "cookie_name")
			testhelpers.AssertEqual(t, signed.secret, tt.want)
		})
	}
}

func Test_Signed_WriteAndRead(t *testing.T) {
	signed := NewSigned("4366d3f57f71049774c039609100ea220467062dfa6eeed93a939629c173ad5d", "cookie_name")
	recorder := httptest.NewRecorder()

	valueToWrite := "cookie_value"
	signed.Write(recorder, valueToWrite, nil)

	// Check if the cookie was set correctly in the response
	cookies := recorder.Result().Cookies()
	testhelpers.AssertEqual(t, len(cookies), 1)

	cookie := cookies[0]
	testhelpers.AssertEqual(t, cookie.Name, "cookie_name")

	// Read the cookie from a mock request
	req := httptest.NewRequest("GET", "/", nil)
	req.AddCookie(cookie)

	valueRead, err := signed.Read(req)
	testhelpers.AssertNoError(t, err)
	testhelpers.AssertEqual(t, valueRead, valueToWrite)
}

func Test_Signed_encode(t *testing.T) {
	signed := NewSigned("17e739297ecfb19eba43c43feda07e1d33f35dee792f20f279b468ee4399e406", "user")
	got := signed.encode("6ae59bb0-bb86-4943-86db-103f632103eg")
	t.Log(got)
}
