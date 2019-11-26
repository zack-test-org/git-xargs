package collections

// Return a new slice that has the contents of the given one in reverse order
func ReverseSlice(slice []string) []string {
	var out []string
	for i := len(slice) - 1; i >= 0; i-- {
		out = append(out, slice[i])
	}
	return out
}

// Returns true if slice b is a subset of slice a.
//
// Examples:
//
// IsSubSetOf([]string{"foo", "bar"}, []string{"foo", "bar"}) // Returns true
// IsSubSetOf([]string{"foo", "bar"}, []string{"foo"})        // Returns true
// IsSubSetOf([]string{"foo", "bar"}, []string{"baz"})        // Returns false
// IsSubSetOf([]string{"foo", "bar"}, []string{"foo", "baz"}) // Returns false
func IsSubsetOf(a []string, b []string) bool {
	if len(b) > len(a) {
		return false
	}

	for i := 0; i < len(b); i++ {
		if a[i] != b[i] {
			return false
		}
	}

	return true
}