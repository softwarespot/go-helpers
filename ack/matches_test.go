package ack

import (
	"os"
	"strings"
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_Print(t *testing.T) {
	multiLines := `test
test
example example
test
example example
example example
test
test
test
test test
test
test
test
test
example example
example example
example example
`
	r := strings.NewReader(multiLines)
	ms, err := Find(r, "test", FindOptions{
		UseCaseSensitive: false,
		UseRegExp:        false,
		MaxCount:         0,
		BeforeContext:    0,
		AfterContext:     0,
		BufferSize:       0,
	})
	testhelpers.AssertNoError(t, err)

	err = ms.Print(os.Stdout, "test.txt", PrintOptions{
		LocationsWithMatches:    false,
		LocationsWithoutMatches: false,
		CountsOnly:              false,
		IsPiped:                 false,
		NoColor:                 false,
	})
	testhelpers.AssertNoError(t, err)
}
