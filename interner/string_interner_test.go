package interner

import (
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_NewStringInterner(t *testing.T) {
	strInterner := NewStringInterner()

	// Should intern strings and return an index
	testhelpers.AssertEqual(t, strInterner.Intern("test-1"), 0)
	testhelpers.AssertEqual(t, strInterner.Intern("test-2"), 1)
	testhelpers.AssertEqual(t, strInterner.Intern("test-3"), 2)

	// Should return the same index when the same string value is interned
	testhelpers.AssertEqual(t, strInterner.Intern("test-1"), 0)

	// Should return an empty string value when the index doesn't exist
	s := strInterner.Resolve(99)
	testhelpers.AssertEqual(t, s, "")

	// Should return the interned string value when the index exists
	s = strInterner.Resolve(0)
	testhelpers.AssertEqual(t, s, "test-1")

	s = strInterner.Resolve(1)
	testhelpers.AssertEqual(t, s, "test-2")

	// Should return all interned string values
	testhelpers.AssertEqual(t, strInterner.Values(), []string{"test-1", "test-2", "test-3"})
}

func Test_NewStringInternerOnly(t *testing.T) {
	strInterner := NewStringInternerOnly()

	// Should intern strings and return an index
	testhelpers.AssertEqual(t, strInterner.Intern("test-1"), 0)
	testhelpers.AssertEqual(t, strInterner.Intern("test-2"), 1)
	testhelpers.AssertEqual(t, strInterner.Intern("test-3"), 2)

	// Should return the same index when the same string value is interned
	testhelpers.AssertEqual(t, strInterner.Intern("test-1"), 0)
}
