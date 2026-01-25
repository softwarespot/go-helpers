package helpers

// SliceAt returns the element at the specified index in the slice.
// It supports negative indexing, allowing you to access elements from the end of the slice.
// If the index is out of bounds (either negative or greater than the length of the slice),
// it returns the zero value of the element type
//
// Example usage:
//
//	slice := []int{10, 20, 30, 40, 50}
//	value1 := SliceAt(slice, 2)    // Returns 30
//	value2 := SliceAt(slice, -1)   // Returns 50
//	value3 := SliceAt(slice, 5)    // Returns 0 (zero value for int)
//	value4 := SliceAt(slice, -6)   // Returns 0 (zero value for int)
func SliceAt[S ~[]E, E any](s S, idx int) E {
	if idx < 0 {
		idx += len(s)
	}
	if idx < 0 || idx >= len(s) {
		var v E
		return v
	}
	return s[idx]
}
