package set

import "iter"

// Examples:
// See URL: https://pkg.go.dev/github.com/hashicorp/go-set
// See URL: https://github.com/deckarep/golang-set

// Set is a generic set implementation using a map
type Set[T comparable] map[T]struct{}

// New creates a new empty set
func New[T comparable]() Set[T] {
	return make(Set[T], 256)
}

// NewFromValues creates a new set from the given values
func NewFromValues[T comparable](vs ...T) Set[T] {
	s := New[T]()
	for _, v := range vs {
		s.Add(v)
	}
	return s
}

// Add returns true when the value is added; otherwise, false when it already exists in the set
func (s Set[T]) Add(v T) bool {
	if s.Has(v) {
		return false
	}

	s[v] = struct{}{}
	return true
}

// Has checks if the value exists in the set
// Returns true if the value exists; otherwise, false
func (s Set[T]) Has(v T) bool {
	_, ok := s[v]
	return ok
}

// Iter returns an iterator over the set values
func (s Set[T]) Iter() iter.Seq[T] {
	return func(yield func(T) bool) {
		for k := range s {
			if !yield(k) {
				return
			}
		}
	}
}

// Values returns a slice of all values in the set
func (s Set[T]) Values() []T {
	vs := make([]T, 0, s.Size())
	for v := range s {
		vs = append(vs, v)
	}
	return vs
}

// Size returns the number of value in the set
func (s Set[T]) Size() int {
	return len(s)
}

// Delete removes the value from the set if it exists.
// Returns true if the value was deleted; otherwise, false
func (s Set[T]) Delete(v T) bool {
	if !s.Has(v) {
		return false
	}

	delete(s, v)
	return true
}

// Clear removes all values from the set
func (s Set[T]) Clear() {
	clear(s)
}
