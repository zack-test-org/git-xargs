package main

import (
	"io/ioutil"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEmpty(t *testing.T) {
	t.Parallel()

	releaseNote, err := parseReleaseNoteBody("")
	assert.NoError(t, err)
	assert.Equal(t, RenderReleaseNote(releaseNote), readFileAsString(t, "test_assets/empty.md"))
}

func TestParseFull(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := parseReleaseNoteBody(fullNote)
	assert.NoError(t, err)
	assert.Equal(t, RenderReleaseNote(releaseNote), fullNote)
}

func TestAppendModulesAffected(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := parseReleaseNoteBody(fullNote)
	assert.NoError(t, err)

	releaseNote = appendModulesAffected(releaseNote, "new-module")
	assert.Equal(t, RenderReleaseNote(releaseNote), readFileAsString(t, "test_assets/full_with_new_module.md"))
}

func TestAppendModulesDedups(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := parseReleaseNoteBody(fullNote)
	assert.NoError(t, err)

	releaseNote = appendModulesAffected(releaseNote, "eks-cluster")
	assert.Equal(t, RenderReleaseNote(releaseNote), readFileAsString(t, "test_assets/full.md"))
}

func TestAppendDescription(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := parseReleaseNoteBody(fullNote)
	assert.NoError(t, err)

	releaseNote = appendDescription(releaseNote, "TODO: Pull Request Title")
	assert.Equal(t, RenderReleaseNote(releaseNote), readFileAsString(t, "test_assets/full_with_new_description.md"))
}

func TestAppendRelatedLink(t *testing.T) {
	t.Parallel()

	fullNote := readFileAsString(t, "test_assets/full.md")
	releaseNote, err := parseReleaseNoteBody(fullNote)
	assert.NoError(t, err)

	releaseNote = appendRelatedLink(releaseNote, "https://github.com/gruntwork-io/package-k8s")
	assert.Equal(t, RenderReleaseNote(releaseNote), readFileAsString(t, "test_assets/full_with_new_link.md"))
}

func readFileAsString(t *testing.T, path string) string {
	rawData, err := ioutil.ReadFile(path)
	require.NoError(t, err)
	return string(rawData)
}
