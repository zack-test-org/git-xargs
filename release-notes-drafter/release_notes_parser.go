package main

import (
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/sirupsen/logrus"

	// This is a nonintuitive name, but this package is a powerful markdown parser/renderer that can parse markdown into
	// a tree of nodes.
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// MarkdownExtensions are blackfriday options that are compatible with github flavored markdown. This is pulled from
// https://github.com/jsternberg/markdownfmt/blob/master/markdown/main.go#L560
const MarkdownExtensions = blackfriday.NoIntraEmphasis |
	blackfriday.Tables |
	blackfriday.FencedCode |
	blackfriday.Autolink |
	blackfriday.Strikethrough |
	blackfriday.SpaceHeadings |
	blackfriday.NoEmptyLineBeforeBlock

// parseReleaseNoteBody takes the release note body as a string and parses it into the ReleaseNote struct by leveraging
// the known markdown structure of the release notes. If the body is an empty string, this will create an empty
// ReleaseNote with each section having the proper heading.
func parseReleaseNoteBody(logger *logrus.Entry, releaseNoteBody string) (ReleaseNote, error) {
	if releaseNoteBody == "" {
		newNote := ReleaseNote{
			ModulesAffected: Section{Heading: ModulesAffectedHeading},
			Description:     Section{Heading: DescriptionHeading},
			RelatedLinks:    Section{Heading: RelatedLinksHeading},
		}
		return newNote, nil
	}
	return parseFromMarkdown(logger, releaseNoteBody)
}

// parseFromMarkdown takes the release note body string and parses it, assuming it is in markdown format.
// The assumed format of the markdown is:
//
// Modules affected
// ----------------
// - `module-one`
// - `module-two`
//
// Description
// -----------
// Description preamble
// - Description of change one.
// - Description of change two.
//
// Related links
// -------------
// - Link to PR #1
// - Link to PR #2
func parseFromMarkdown(logger *logrus.Entry, body string) (ReleaseNote, error) {
	releaseNote := ReleaseNote{}

	option := blackfriday.WithExtensions(MarkdownExtensions)
	md := blackfriday.New(option)
	mdTree := md.Parse([]byte(body))
	if mdTree.Type != blackfriday.Document {
		return releaseNote, errors.WithStackTrace(ReleaseNoteParsingError{body})
	}

	// blackfriday will render the markdown body as a tree, rooted on the Document node. The Document node acts as the
	// root, with all the body contents branching off of it as direct descendents.
	// The idea here is to parse section by section and add it to the corresponding release note section.
	// See docs on parseSection for more info.
	currentNode := mdTree.FirstChild
	for currentNode != nil {
		section, node, err := parseSection(logger, currentNode)
		if err != nil {
			return releaseNote, err
		}
		updated := updateReleaseNoteWithSection(&releaseNote, section)
		if !updated {
			return releaseNote, errors.WithStackTrace(UnknownHeadingError{section.Heading, body})
		}
		currentNode = node
	}

	return releaseNote, nil
}

// updateReleaseNoteWithSection will set the appropriate release note section based on the section heading.
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

// parseSection will take a tree node and parse it into a release note section based on the following heuristic:
// blackfriday will render the markdown body as a tree, rooted on the Document node. The Document node acts as the
// root, with all the body contents branching off of it as direct descendents.
// This means that from the root document, each heading will be in the same level.
// Therefore, we parse each section by walking the direct descendents in order, assuming everything after a heading
// node to belongs to that heading as a section. Then, we parse the contents of the section into the prescribed format
// of:
// HEADING
// PREAMBLE
// DETAILS_LIST
// Where:
// - HEADING is the section heading.
// - PREAMBLE is everything before the detailed list of info for the section
// - DETAILS_LIST is the list of details for the section (e.g list of modules for "Modules affected" section).
// To ensure the markdown renders correctly, the list items and the preamble are stored as blackfriday nodes directly.
// This returns which node we parsed until in the tree, which is either the end of the tree, or the next section
// heading.
func parseSection(logger *logrus.Entry, node *blackfriday.Node) (Section, *blackfriday.Node, error) {
	section := Section{}

	// Ensure that the right heading was passed
	if node.Type != blackfriday.Heading {
		logger.Errorf("Invoked parseSection with a non heading node")
		return section, nil, errors.WithStackTrace(IncorrectParserError{node.String()})
	}

	section.Heading = string(node.FirstChild.Literal)
	node = node.Next
	for node != nil && node.Type != blackfriday.Heading {
		// Everything that is after the heading, and is not a list is appended to the preamble.
		// TODO: This does not handle multiple lists. This may be problematic if the description section is complex.
		switch node.Type {
		case blackfriday.List:
			list, err := parseList(logger, node)
			if err != nil {
				return section, nil, err
			}
			section.Details = list
		default:
			section.Preamble = append(section.Preamble, node)
		}
		node = node.Next
	}
	return section, node, nil
}

// parseList will take a markdown tree node and parse it as a list entry. This can handle both ordered and unordered
// lists.
func parseList(logger *logrus.Entry, node *blackfriday.Node) (List, error) {
	// Make sure the provided node is of the right type
	if node.Type != blackfriday.List {
		logger.Errorf("Invoked parseList with a non list node")
		return List{}, errors.WithStackTrace(IncorrectParserError{node.String()})
	}

	// Check if list is ordered, and set the appropriate flags.
	list := List{
		IsOrdered: node.ListFlags == blackfriday.ListTypeOrdered,
	}
	// Add all the child nodes to the list object.
	node = node.FirstChild
	for node != nil {
		list.Items = append(list.Items, node)
		node = node.Next
	}
	return list, nil
}
