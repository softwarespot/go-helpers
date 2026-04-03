package testhelpers

import (
	"reflect"
	"testing"
)

// AssertEqual checks if two values are equal. If they are not, it logs using t.Fatalf()
func AssertEqual[T any](t testing.TB, got, correct T) {
	t.Helper()
	if !reflect.DeepEqual(got, correct) {
		t.Fatalf("AssertEqual: expected values to be equal, got:\n%+v\ncorrect:\n%+v", got, correct)
	}
}

// AssertError checks if an error is not nil. If it's nil, it logs using t.Fatalf()
func AssertError(t testing.TB, err error) {
	t.Helper()
	if err == nil {
		t.Fatalf("AssertError: expected an error, got nil")
	}
}

// AssertNoError checks if an error is nil. If it's not nil, it logs using t.Fatalf()
func AssertNoError(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatalf("AssertNoError: expected no error, got %+v", err)
	}
}
