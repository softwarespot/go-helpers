package helpers

// Idea taken from URL: https://github.com/samber/lo/blob/master/slice.go

// SliceUnique returns a new slice containing only the unique elements
// from the provided slice. The order of elements is preserved.
//
// Example:
//
//	helpers.SliceUnique([]int{1, 2, 2, 3, 1, 4, 5}) // Returns []int{1, 2, 3, 4, 5}
//	helpers.SliceUnique([]string{"a", "b", "a"})    // Returns []string{"a", "b"}
func SliceUnique[S ~[]E, E comparable](s S) S {
	return SliceUniqueFunc(s, func(e E) E { return e })
}

// SliceUniqueFunc returns a new slice containing only the unique elements
// determined by the provided keyFunc. This allows uniqueness to be determined
// by a specific property of complex types.
//
// Example:
//
//	type Person struct {
//	    ID   int
//	    Name string
//	}
//
//	people := []Person{{1, "Alice"}, {2, "Bob"}, {1, "Alice_Clone"}}
//	unique := helpers.SliceUniqueFunc(people, func(p Person) int { return p.ID })
//	// Returns []Person{{1, "Alice"}, {2, "Bob"}}
func SliceUniqueFunc[S ~[]E, E any, K comparable](s S, keyFunc func(E) K) S {
	if len(s) == 0 {
		return S{}
	}

	res := make(S, 0, len(s))
	seen := make(map[K]struct{}, len(s))
	for _, v := range s {
		key := keyFunc(v)
		if _, ok := seen[key]; !ok {
			seen[key] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}
