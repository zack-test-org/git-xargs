package collections

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"strings"
	"testing"
)

func TestReverseSlice(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		input    []string
		expected []string
	}{
		{[]string{}, nil},
		{[]string{"a"}, []string{"a"}},
		{[]string{"a", "b"}, []string{"b", "a"}},
		{[]string{"a", "b", "c"}, []string{"c", "b", "a"}},
		{[]string{"a", "b", "c", "d", "e"}, []string{"e", "d", "c", "b", "a"}},
	}

	for _, testCase := range testCases {
		// The following is necessary to make sure testCase's values don't
		// get updated due to concurrency within the scope of t.Run(..) below
		testCase := testCase
		t.Run(strings.Join(testCase.input, "_"), func(t *testing.T) {
			actual := ReverseSlice(testCase.input)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}

func TestIsSubsetOf(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		a        []string
		b        []string
		expected bool
	}{
		{[]string{}, []string{}, true},
		{[]string{"a"}, []string{}, true},
		{[]string{"a", "b"}, []string{}, true},
		{[]string{"a", "b", "c"}, []string{}, true},
		{[]string{"a"}, []string{"a"}, true},
		{[]string{"a", "b"}, []string{"a"}, true},
		{[]string{"a", "b", "c"}, []string{"a"}, true},
		{[]string{"a", "b"}, []string{"a", "b"}, true},
		{[]string{"a", "b", "c"}, []string{"a", "b"}, true},
		{[]string{"a", "b", "c", "d", "e"}, []string{"a", "b"}, true},
		{[]string{"a", "b", "c", "d", "e"}, []string{"a", "b", "c", "d"}, true},
		{[]string{}, []string{"a"}, false},
		{[]string{"a"}, []string{"b"}, false},
		{[]string{"a", "b"}, []string{"b"}, false},
		{[]string{"a", "b"}, []string{"a", "c"}, false},
		{[]string{"a", "b", "c", "d", "e"}, []string{"a", "b", "c", "c", "d", "e"}, false},
	}

	for _, testCase := range testCases {
		// The following is necessary to make sure testCase's values don't
		// get updated due to concurrency within the scope of t.Run(..) below
		testCase := testCase
		testCaseName := fmt.Sprintf("%s__%s", strings.Join(testCase.a, "_"), strings.Join(testCase.b, "_"))
		t.Run(testCaseName, func(t *testing.T) {
			actual := IsSubsetOf(testCase.a, testCase.b)
			assert.Equal(t, testCase.expected, actual)
		})
	}
}
