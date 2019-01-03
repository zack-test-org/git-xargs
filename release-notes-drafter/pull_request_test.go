package main

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestBasicModulesAffectedFromDiff(t *testing.T) {
	t.Parallel()

	diffString := readFileAsString(t, "test_assets/basic.diff")
	modulesAffected, err := extractModulesAffectedFromDiff(diffString)
	require.NoError(t, err)
	require.Equal(t, len(modulesAffected), 1)
	require.Equal(t, modulesAffected[0], "new")
}
