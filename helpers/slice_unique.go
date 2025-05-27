package helpers

// Idea taken from URL: https://github.com/samber/lo/blob/master/slice.go

// SliceUnique returns a new slice containing only the unique elements
// from the provided slice
func SliceUnique[S ~[]E, E comparable](s S) S {
	if len(s) == 0 {
		return nil
	}

	res := make(S, 0, len(s))
	seen := make(map[E]struct{}, len(s))
	for _, v := range s {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			res = append(res, v)
		}
	}
	return res
}
