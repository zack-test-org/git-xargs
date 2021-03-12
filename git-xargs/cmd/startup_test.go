package cmd

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestVerifyDependenciesInstalledReturnsMissingDeps(t *testing.T) {
	depsNobodyHasInstalled := []Dependency{
		{Name: "fozzie-bears-bestest-binary", URL: "https://the-great-over-there/fozzie-bear-install.html"},
	}

	ok, missingDeps := verifyDependenciesInstalled(depsNobodyHasInstalled)

	assert.False(t, ok)

	assert.Equal(t, len(missingDeps), 1)
}
