package helpers

// ToPtr returns a pointer to the provided value
func ToPtr[T any](v T) *T {
	return &v
}

// FromPtr safely dereferences a pointer, returning the zero value if nil
func FromPtr[T any](v *T) T {
	if v == nil {
		var res T
		return res
	}
	return *v
}
