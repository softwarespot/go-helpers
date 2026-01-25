package helpers

import (
	"errors"
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_RecoverFunc(t *testing.T) {
	var errRecovered error
	errUnexepected := errors.New("unexpected panic")
	func() {
		defer RecoverFunc(func(err error) {
			errRecovered = err
		})
		panic(errUnexepected)
	}()
	testhelpers.AssertEqual(t, errors.Unwrap(errRecovered) == errUnexepected, true)
}
