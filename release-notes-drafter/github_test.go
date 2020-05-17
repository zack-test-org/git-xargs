package main

import (
	"testing"

	"github.com/google/go-github/v31/github"
	"github.com/stretchr/testify/assert"
)

func TestBumpPatchVersion(t *testing.T) {
	t.Parallel()

	var testCases = []struct {
		in  string
		out string
	}{
		{"", "v0.0.1"},
		{"v0.0.1", "v0.0.2"},
		{"v1.10.100", "v1.10.101"},
		{"0.0.1", "v0.0.2"},
		// Error cases denoted by empty string for out
		{"this is not a semantic version string", ""},
		{"0.10", ""},
		{"v0", ""},
		{"v0.0.one", ""},
	}

	for _, testCase := range testCases {
		// redefine testCase so that it stays consistent in the run function (since it is defined outside this block).
		testCase := testCase
		t.Run(testCase.in, func(t *testing.T) {
			t.Parallel()

			var mockLastRelease *github.RepositoryRelease
			if testCase.in == "" {
				mockLastRelease = nil
			} else {
				mockLastRelease = &github.RepositoryRelease{
					TagName: github.String(testCase.in),
				}
			}

			bumpedVersion, err := bumpPatchVersion(mockLastRelease)
			if testCase.out == "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, bumpedVersion, testCase.out)
			}
		})
	}
}
