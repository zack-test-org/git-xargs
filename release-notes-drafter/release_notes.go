package main

import (
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// List represents a markdown list, without the ListOpen and ListClose tokens
type List struct {
	IsOrdered bool
	Items     []*blackfriday.Node
}

type Section struct {
	Heading  string
	Preamble []*blackfriday.Node
	Details  List
}

type ReleaseNote struct {
	ModulesAffected Section
	Description     Section
	RelatedLinks    Section
}

const (
	ModulesAffectedHeading = "Modules affected"
	DescriptionHeading     = "Description"
	RelatedLinksHeading    = "Related links"
)

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
