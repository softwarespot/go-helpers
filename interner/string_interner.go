package interner

// StringInterner provides a memory efficient storage for repeated string values by
// mapping each unique string value to a string unique index
type StringInterner struct {
	// Fast lookup for the interned string value to index
	idxByValue map[string]int32

	// Fast lookup for the index to interned string value
	values  []string
	resolve bool
}

func NewStringInterner() *StringInterner {
	return &StringInterner{
		idxByValue: make(map[string]int32, 256),
		values:     make([]string, 0, 256),
		resolve:    true,
	}
}

// NewStringInternerOnly return a string interner that does not resolve the string values.
// This is useful when you only need to intern strings and do not need to resolve them back to their original values.
// This can save memory and improve performance in scenarios where the string values are not needed
func NewStringInternerOnly() *StringInterner {
	return &StringInterner{
		idxByValue: make(map[string]int32, 256),
		values:     make([]string, 0, 256),
		resolve:    false,
	}
}

// Intern returns the index of the interned string value, returning the same index if the string value
// has already been interned
func (si *StringInterner) Intern(v string) int32 {
	if idx, ok := si.idxByValue[v]; ok {
		return idx
	}

	idx := int32(len(si.idxByValue))
	si.idxByValue[v] = idx

	if si.resolve {
		si.values = append(si.values, v)
	}
	return idx
}

// Resolve returns the interned string value for the provided index.
// NOTE: This panics if the interner was created with "NewStringInternerOnly"
func (si *StringInterner) Resolve(idx int32) string {
	if !si.resolve {
		panic("interner.Resolve: called on a StringInterner that was created with NewStringInternerOnly")
	}
	if idx < 0 || idx >= int32(len(si.values)) {
		return ""
	}
	return si.values[idx]
}

// Values returns a slice of all interned string values.
// NOTE: This panics if the interner was created with "NewStringInternerOnly"
func (si *StringInterner) Values() []string {
	if !si.resolve {
		panic("interner.Values: called on a StringInterner that was created with NewStringInternerOnly")
	}
	return si.values
}
