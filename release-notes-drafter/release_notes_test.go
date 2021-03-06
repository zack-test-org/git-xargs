package main

import (
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDedupLinesPreservesTemplate(t *testing.T) {
	t.Parallel()

	lines := strings.Split(ReleaseNoteTemplate, "\n")
	assert.Equal(t, dedupLines(lines, DedupWhitelist), lines)
}

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

func TestAddContributor(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := addContributor(fullNote, "grunty")
	assert.NoError(t, err)
	assert.Equal(t, releaseNote, readFileAsString(t, "test_assets/full_with_new_contributor.md"))
}

func TestAddModulesAffectedDedups(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := addModulesAffected(fullNote, []string{"kubergrunt"})
	assert.NoError(t, err)
	assert.Equal(t, releaseNote, readFileAsString(t, "test_assets/full.md"))
}

func TestAddDescriptionDedups(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := addDescription(fullNote, "A faux description")
	assert.NoError(t, err)
	assert.Equal(t, releaseNote, readFileAsString(t, "test_assets/full.md"))
}

func TestAddRelatedLinkDedups(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := addRelatedLink(fullNote, "[PR](https://github.com/gruntwork-io/package-k8s)")
	assert.NoError(t, err)
	assert.Equal(t, releaseNote, readFileAsString(t, "test_assets/full.md"))
}

func readFileAsString(t *testing.T, path string) string {
	rawData, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return string(rawData)
}
