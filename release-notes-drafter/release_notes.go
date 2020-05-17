package main

import (
	"fmt"
	"strings"

	"github.com/gruntwork-io/gruntwork-cli/collections"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

// Markers that denote where the next information should be inserted in the release notes.
const (
	ModulesAffectedMarker = "<!-- RELEASE_NOTES_DRAFTER_MARKER_MODULES_AFFECTED_NEXT -->"
	DescriptionMarker     = "<!-- RELEASE_NOTES_DRAFTER_MARKER_DESCRIPTIONS_NEXT -->"
	RelatedLinksMarker    = "<!-- RELEASE_NOTES_DRAFTER_MARKER_RELATED_LINKS_NEXT -->"
)

// We don't want to dedup these markers that are used for significant whitespace in comment sections
var DedupWhitelist = []string{"--", "-->", "<!--", ""}

// findMarker will search the list of strings representing each line in the release note body for the given marker and
// return the index of the marker. This will return -1 if it could not find the index.
func findMarker(bodyLines []string, marker string) int {
	for idx, item := range bodyLines {
		if strings.TrimSpace(item) == marker {
			return idx
		}
	}
	return -1
}

// addModulesAffected will search the release note body for the marker where the next modules affected should be
// inserted, and insert them in as inline code list items.
func addModulesAffected(releaseNoteBody string, modulesAffected []string) (string, error) {
	var err error
	for _, newModuleAffected := range modulesAffected {
		releaseNoteBody, err = findMarkerAndInsertLine(releaseNoteBody, ModulesAffectedMarker, fmt.Sprintf("- `%s`", newModuleAffected))
		if err != nil {
			return releaseNoteBody, err
		}
	}
	return releaseNoteBody, nil
}

// addDescription will search the release note body for the marker where the next description should be inserted, and
// inserts it in verbatim as a list item.
func addDescription(releaseNoteBody string, description string) (string, error) {
	return findMarkerAndInsertLine(releaseNoteBody, DescriptionMarker, fmt.Sprintf("- %s", description))
}

// addRelatedLink will search the release note body for the marker where the next link should be inserted, and inserts
// it in verbatim as a list item.
func addRelatedLink(releaseNoteBody string, relatedLink string) (string, error) {
	return findMarkerAndInsertLine(releaseNoteBody, RelatedLinksMarker, fmt.Sprintf("- %s", relatedLink))
}

// findMarkerAndInsertLine will insert a new line where the given text marker is.
func findMarkerAndInsertLine(releaseNoteBody string, marker string, line string) (string, error) {
	bodyLines := strings.Split(releaseNoteBody, "\n")
	markerIdx := findMarker(bodyLines, marker)
	if markerIdx == -1 {
		return releaseNoteBody, errors.WithStackTrace(MissingMarkerError{marker, releaseNoteBody})
	}
	bodyLines = insertLine(bodyLines, markerIdx, line)
	bodyLines = dedupLines(bodyLines, DedupWhitelist)
	return strings.Join(bodyLines, "\n"), nil
}

// insertLine will insert the provided new line to the list of lines at the given index.
// https://github.com/golang/go/wiki/SliceTricks#insert
func insertLine(lines []string, idx int, line string) []string {
	return append(lines[:idx], append([]string{line}, lines[idx:]...)...)
}

// dedupLines will dedup lines that are equivalent after trimming spaces. Will always add lines that match the
// whitelist.
func dedupLines(lines []string, whitelistLines []string) []string {
	seen := map[string]bool{}
	outLines := []string{}
	for _, line := range lines {
		trimmedLine := strings.TrimSpace(line)
		_, hasSeen := seen[trimmedLine]
		if !hasSeen || collections.ListContainsElement(whitelistLines, trimmedLine) {
			outLines = append(outLines, line)
			seen[trimmedLine] = true
		}
	}
	return outLines
}

// ReleaseNoteTemplate represents the template to use for starting a new release notes. This has information for the
// maintainer about how to update it, as well as information
const ReleaseNoteTemplate = `<!--
  -- This is autogenerated from the release notes drafter. When updating, be sure to double check some of the changes
  -- before publishing.
  -- Note that there are markers for the release notes drafter as comments. DO NOT REMOVE THEM. They will not show up in
  -- the preview and is harmless to keep, but harmful to remove as it is used to guide the drafter on where the next
  -- information should be inserted.
  -->

## Modules affected

<!-- The list of modules that have been touched since the last release.
  --
  -- The autogenerator will choose to do a patch release. However, check if the changes in the following modules are
  -- backwards compatible, and update the release number if it is backwards incompatible.
  --
  -- The following kinds of changes would constitute a backwards incompatible change:
  -- * In Terraform code: add a new variable without a default, rename or remove an existing variable, remove or rename
  --   an output, remove or rename a resource.
  -- * In Bash and Go code: add a new parameter without a default, rename or remove an existing parameter, fundamentally
  --   change what the code does.
  -->

<!-- RELEASE_NOTES_DRAFTER_MARKER_MODULES_AFFECTED_NEXT -->


## Description

<!-- A description of the changes made in this release. Be sure to update any TODOs. -->

<!-- RELEASE_NOTES_DRAFTER_MARKER_DESCRIPTIONS_NEXT -->


## Related links

<!-- Links to each PR or issue that are being addressed in this release. The drafter will autoinclude each merged PR. -->

<!-- RELEASE_NOTES_DRAFTER_MARKER_RELATED_LINKS_NEXT -->

`
