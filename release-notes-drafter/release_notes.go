package main

import (
	// This is a nonintuitive name, but this package is a powerful markdown parser/renderer that can parse markdown into
	// a tree of nodes.
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// List represents a markdown list, without the ListOpen and ListClose tokens
// If it is an ordered list, IsOrdered will be true.
type List struct {
	IsOrdered bool
	Items     []*blackfriday.Node
}

// Section represents a release note section. A section is composed of:
// - A heading
// - A preamble, which is text right after the heading before a bullet list of items
// - A bullet list of items
type Section struct {
	Heading  string
	Preamble []*blackfriday.Node
	Details  List
}

// ReleaseNote represents a release note in the Gruntwork format. Any changes to the format should be made here.
type ReleaseNote struct {
	ModulesAffected Section
	Description     Section
	RelatedLinks    Section
}

// This is the list of known headings for a release note format
const (
	ModulesAffectedHeading = "Modules affected"
	DescriptionHeading     = "Description"
	RelatedLinksHeading    = "Related links"
)

// appendModulesAffected will take a release note and append the provided module affected (as a string) to the release
// note as a line item wrapped in inline code quotes.
func appendModulesAffected(releaseNote ReleaseNote, moduleAffected string) ReleaseNote {
	itemNode := blackfriday.NewNode(blackfriday.Item)
	paragraphNode := blackfriday.NewNode(blackfriday.Paragraph)
	textNode := blackfriday.NewNode(blackfriday.Text)
	codeNode := blackfriday.NewNode(blackfriday.Code)
	codeNode.Literal = []byte(moduleAffected)
	paragraphNode.AppendChild(textNode)
	paragraphNode.AppendChild(codeNode)
	itemNode.AppendChild(paragraphNode)

	releaseNote.ModulesAffected.Details.Items = append(releaseNote.ModulesAffected.Details.Items, itemNode)
	return releaseNote
}

// appendDescription will take a release note and append the provided description (as a string) to the release note as
// a simple text link item.
func appendDescription(releaseNote ReleaseNote, description string) ReleaseNote {
	itemNode := blackfriday.NewNode(blackfriday.Item)
	paragraphNode := blackfriday.NewNode(blackfriday.Paragraph)
	textNode := blackfriday.NewNode(blackfriday.Text)
	textNode.Literal = []byte(description)
	paragraphNode.AppendChild(textNode)
	itemNode.AppendChild(paragraphNode)

	releaseNote.Description.Details.Items = append(releaseNote.Description.Details.Items, itemNode)
	return releaseNote
}

// appendRelatedLink will take a release note and append the provided link URL (as a string) to the release note as a
// link node.
func appendRelatedLink(releaseNote ReleaseNote, url string) ReleaseNote {
	itemNode := blackfriday.NewNode(blackfriday.Item)
	paragraphNode := blackfriday.NewNode(blackfriday.Paragraph)
	linkNode := blackfriday.NewNode(blackfriday.Link)
	linkNode.Destination = []byte(url)
	linkTextNode := blackfriday.NewNode(blackfriday.Text)
	linkTextNode.Literal = []byte(url)
	linkNode.AppendChild(linkTextNode)
	paragraphNode.AppendChild(linkNode)
	itemNode.AppendChild(paragraphNode)

	releaseNote.RelatedLinks.Details.Items = append(releaseNote.RelatedLinks.Details.Items, itemNode)
	return releaseNote
}
