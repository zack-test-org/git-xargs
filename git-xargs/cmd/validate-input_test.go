package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnsureValidOptionsPassedRejectsEmptySelectors(t *testing.T) {

	ok := ensureValidOptionsPassed("", "")

	assert.False(t, ok)
}

func TestEnsureValidOptionsPassedAcceptsValidGithubOrg(t *testing.T) {

	ok := ensureValidOptionsPassed("", "gruntwork-io")

	assert.True(t, ok)
}
