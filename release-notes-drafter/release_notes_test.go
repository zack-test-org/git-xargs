package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddModulesAffected(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := addModulesAffected(fullNote, []string{"new-module", "another-module"})
	assert.NoError(t, err)
	assert.Equal(t, releaseNote, readFileAsString(t, "test_assets/full_with_new_module.md"))
}

func TestAddDescription(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := addDescription(fullNote, "TODO: Pull request title")
	assert.NoError(t, err)
	assert.Equal(t, releaseNote, readFileAsString(t, "test_assets/full_with_new_description.md"))
}

func TestAddRelatedLink(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := addRelatedLink(fullNote, "https://github.com/gruntwork-io/package-k8s")
	assert.NoError(t, err)
	assert.Equal(t, releaseNote, readFileAsString(t, "test_assets/full_with_new_link.md"))
}

func readFileAsString(t *testing.T, path string) string {
	rawData, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return string(rawData)
}
