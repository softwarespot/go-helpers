package set

import (
	"slices"
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_New(t *testing.T) {
	s := NewFromValues[int]()
	testhelpers.AssertEqual(t, len(s), 0)
	testhelpers.AssertEqual(t, s.Has(1), false)
	testhelpers.AssertEqual(t, s.Has(2), false)
	testhelpers.AssertEqual(t, s.Has(3), false)
	testhelpers.AssertEqual(t, s.Has(4), false)
	testhelpers.AssertEqual(t, s.Values(), []int{})

	// Should add 3 values
	s = NewFromValues(1, 2, 3)
	testhelpers.AssertEqual(t, len(s), 3)

	// Should check if the values are in the set
	testhelpers.AssertEqual(t, s.Has(1), true)
	testhelpers.AssertEqual(t, s.Has(2), true)
	testhelpers.AssertEqual(t, s.Has(3), true)
	testhelpers.AssertEqual(t, s.Has(4), false)

	// Should check if the value is in the set or add
	testhelpers.AssertEqual(t, s.Add(1), false)
	testhelpers.AssertEqual(t, s.Add(2), false)
	testhelpers.AssertEqual(t, s.Add(5), true)
	testhelpers.AssertEqual(t, s.Has(5), true)
	testhelpers.AssertEqual(t, s.Add(5), false)

	// Should get the values
	vs := s.Values()
	slices.Sort(vs)
	testhelpers.AssertEqual(t, vs, []int{1, 2, 3, 5})

	// Should delete a value
	testhelpers.AssertEqual(t, s.Delete(2), true)
	testhelpers.AssertEqual(t, s.Delete(2), false)

	testhelpers.AssertEqual(t, s.Has(1), true)
	testhelpers.AssertEqual(t, s.Has(2), false)
	testhelpers.AssertEqual(t, s.Has(3), true)
	testhelpers.AssertEqual(t, s.Has(4), false)

	vs = s.Values()
	slices.Sort(vs)
	testhelpers.AssertEqual(t, vs, []int{1, 3, 5})

	// Should iterate over the values and break
	vs = nil
	for v := range s.Iter() {
		if len(vs) == 2 {
			break
		}
		vs = append(vs, v)
	}
	testhelpers.AssertEqual(t, len(vs), 2)

	// Should clear the set
	s.Clear()
	testhelpers.AssertEqual(t, len(s), 0)
	testhelpers.AssertEqual(t, s.Has(1), false)
	testhelpers.AssertEqual(t, s.Has(2), false)
	testhelpers.AssertEqual(t, s.Has(3), false)
	testhelpers.AssertEqual(t, s.Has(4), false)
	testhelpers.AssertEqual(t, s.Values(), []int{})
}
