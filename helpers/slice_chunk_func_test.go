package helpers

import (
	"errors"
	"testing"

	testhelpers "github.com/softwarespot/go-helpers/test-helpers"
)

func Test_SliceChunkFunc(t *testing.T) {
	// Should chunk successfully when size if less than the slice length
	var numsBatched [][]int
	err := SliceChunkFunc([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 4, func(nums []int) error {
		numsBatched = append(numsBatched, nums)
		return nil
	})
	testhelpers.AssertNoError(t, err)
	testhelpers.AssertEqual(t, numsBatched, [][]int{
		{1, 2, 3, 4},
		{5, 6, 7, 8},
		{9, 10},
	})

	// Should chunk successfully when size is 1
	numsBatched = nil
	err = SliceChunkFunc([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 1, func(nums []int) error {
		numsBatched = append(numsBatched, nums)
		return nil
	})
	testhelpers.AssertNoError(t, err)
	testhelpers.AssertEqual(t, numsBatched, [][]int{
		{1},
		{2},
		{3},
		{4},
		{5},
		{6},
		{7},
		{8},
		{9},
		{10},
	})

	// Should chunk successfully when size if greater than the slice length
	numsBatched = nil
	err = SliceChunkFunc([]int{1, 2, 3, 4, 5}, 10, func(nums []int) error {
		numsBatched = append(numsBatched, nums)
		return nil
	})
	testhelpers.AssertNoError(t, err)
	testhelpers.AssertEqual(t, numsBatched, [][]int{
		{1, 2, 3, 4, 5},
	})

	// Should chunk successfully when size if equal to the slice length
	numsBatched = nil
	err = SliceChunkFunc([]int{1, 2, 3}, 3, func(nums []int) error {
		numsBatched = append(numsBatched, nums)
		return nil
	})
	testhelpers.AssertNoError(t, err)
	testhelpers.AssertEqual(t, numsBatched, [][]int{
		{1, 2, 3},
	})

	// Should chunk successfully with an empty slice
	numsBatched = nil
	err = SliceChunkFunc([]int{}, 4, func(nums []int) error {
		numsBatched = append(numsBatched, nums)
		return nil
	})
	testhelpers.AssertNoError(t, err)
	testhelpers.AssertEqual(t, numsBatched, nil)

	// Should chunk successfully with a nil slice
	numsBatched = nil
	err = SliceChunkFunc(nil, 4, func(nums []int) error {
		numsBatched = append(numsBatched, nums)
		return nil
	})
	testhelpers.AssertNoError(t, err)
	testhelpers.AssertEqual(t, numsBatched, nil)

	// Should stop on error when call to first function returns error
	var strsBatched [][]string
	wantErr := errors.New("unexpected error")
	err = SliceChunkFunc([]string{"A", "B", "C", "D", "E", "F", "G"}, 3, func(strs []string) error {
		strsBatched = append(strsBatched, strs)
		return wantErr
	})
	testhelpers.AssertError(t, err)
	testhelpers.AssertEqual(t, err, wantErr)
	testhelpers.AssertEqual(t, strsBatched, [][]string{
		{"A", "B", "C"},
	})

	// Should stop on error when call to last function returns error
	var boolsBatched [][]bool
	err = SliceChunkFunc([]bool{true, false, true}, 2, func(bools []bool) error {
		boolsBatched = append(boolsBatched, bools)
		if len(boolsBatched) == 2 {
			return wantErr
		}
		return nil
	})
	testhelpers.AssertError(t, err)
	testhelpers.AssertEqual(t, err, wantErr)
	testhelpers.AssertEqual(t, boolsBatched, [][]bool{
		{true, false},
		{true},
	})

	// Should stop on error when called with size 0
	err = SliceChunkFunc([]int{1, 2, 3, 4, 5, 6, 7, 8, 9, 10}, 0, func(nums []int) error {
		return nil
	})
	testhelpers.AssertError(t, err)
}
