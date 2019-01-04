package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetModuleString(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		in  string
		out string
	}{
		{"this/is/not/a/module/path", ""},
		{"modules/this/is/a/module/path", "this"},
		{"modules/README.md", ""},
	}
	for _, testCase := range testCases {
		// Redefine testCase so that it is scoped within this block. This prevents using an outdated version of the
		// variable, which gets updated outside the block in the for loop.
		t.Run(testCase.in, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, getModuleString(testCase.in), testCase.out)
		})
	}
}
