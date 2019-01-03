package main

import (
	"github.com/gruntwork-io/gruntwork-cli/errors"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

const MarkdownExtensions = blackfriday.NoIntraEmphasis |
	blackfriday.Tables |
	blackfriday.FencedCode |
	blackfriday.Autolink |
	blackfriday.Strikethrough |
	blackfriday.SpaceHeadings |
	blackfriday.NoEmptyLineBeforeBlock

func parseReleaseNoteBody(releaseNoteBody string) (ReleaseNote, error) {
	if releaseNoteBody == "" {
		newNote := ReleaseNote{
			ModulesAffected: Section{Heading: ModulesAffectedHeading},
			Description:     Section{Heading: DescriptionHeading},
			RelatedLinks:    Section{Heading: RelatedLinksHeading},
		}
		return newNote, nil
	}
	return parseFromMarkdown(releaseNoteBody)
}

func parseFromMarkdown(body string) (ReleaseNote, error) {
	releaseNote := ReleaseNote{}
	option := blackfriday.WithExtensions(MarkdownExtensions)
	md := blackfriday.New(option)
	mdTree := md.Parse([]byte(body))
	if mdTree.Type != blackfriday.Document {
		return releaseNote, errors.WithStackTrace(ReleaseNoteParsingError{body})
	}
	currentNode := mdTree.FirstChild
	for currentNode != nil {
		section, node := parseSection(currentNode)
		updated := updateReleaseNoteWithSection(&releaseNote, section)
		if !updated {
			return releaseNote, errors.WithStackTrace(UnknownHeadingError{section.Heading, body})
		}
		currentNode = node
	}
	return releaseNote, nil
}

func updateReleaseNoteWithSection(releaseNote *ReleaseNote, section Section) bool {
	switch section.Heading {
	case ModulesAffectedHeading:
		releaseNote.ModulesAffected = section
	case DescriptionHeading:
		releaseNote.Description = section
	case RelatedLinksHeading:
		releaseNote.RelatedLinks = section
	default:
		return false
	}
	return true
}

func parseSection(node *blackfriday.Node) (Section, *blackfriday.Node) {
	section := Section{}
	if node.Type != blackfriday.Heading {
		return section, nil
	}
	section.Heading = string(node.FirstChild.Literal)
	node = node.Next
	for node != nil && node.Type != blackfriday.Heading {
		switch node.Type {
		case blackfriday.List:
			list := parseList(node)
			section.Details = list
		default:
			section.Preamble = append(section.Preamble, node)
		}
		node = node.Next
	}
	return section, node
}

func parseList(node *blackfriday.Node) List {
	list := List{
		IsOrdered: node.ListFlags == blackfriday.ListTypeOrdered,
	}
	node = node.FirstChild
	for node != nil {
		list.Items = append(list.Items, node)
		node = node.Next
	}
	return list
}

func clearPointers(node *blackfriday.Node) {
	node.Prev = nil
	node.Next = nil
	node.Parent = nil
}

/*
(dlv) print mdTree.Type
github.com/gruntwork-io/prototypes/release-notes-drafter/vendor/gopkg.in/russross/blackfriday.v2.Document
(dlv) print mdTree.FirstChild.Type
github.com/gruntwork-io/prototypes/release-notes-drafter/vendor/gopkg.in/russross/blackfriday.v2.Heading
(dlv) print mdTree.FirstChild.Next.Type
github.com/gruntwork-io/prototypes/release-notes-drafter/vendor/gopkg.in/russross/blackfriday.v2.Paragraph
(dlv) print mdTree.FirstChild.FirstChild.Type
github.com/gruntwork-io/prototypes/release-notes-drafter/vendor/gopkg.in/russross/blackfriday.v2.Text
*/
