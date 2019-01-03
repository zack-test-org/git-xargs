package main

import (
	"bytes"

	// We are using jsternberg's fork of markdownfmt because of https://github.com/shurcooL/markdownfmt/pull/40
	// We need blackfriday v2 because of https://github.com/russross/blackfriday#known-issue-with-dep (and superior API)
	"github.com/jsternberg/markdownfmt/markdown"

	// This is a nonintuitive name, but this package is a powerful markdown parser/renderer that can parse markdown into
	// a tree of nodes.
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

// RenderReleaseNote will take the release note object and render it back to a markdown string in the format:
/*
Modules affected
----------------

- `module-one`

- `module-two`

Description
-----------

Description preamble

- Description of change one.

- Description of change two.

Related links
-------------

- Link to PR #1

- Link to PR #2
*/
// Note that the final rendered string has all the extra whitespaces, which are prescribed by the markdownfmt tool.
func RenderReleaseNote(releaseNote ReleaseNote) string {
	// First render each section into blackfriday nodes, and then render the tree as a string.
	rootNode := blackfriday.NewNode(blackfriday.Document)
	for _, node := range renderSectionAsNodes(releaseNote.ModulesAffected) {
		rootNode.AppendChild(node)
	}
	for _, node := range renderSectionAsNodes(releaseNote.Description) {
		rootNode.AppendChild(node)
	}
	for _, node := range renderSectionAsNodes(releaseNote.RelatedLinks) {
		rootNode.AppendChild(node)
	}

	return nodeAsString(rootNode)
}

// renderSectionAsNodes will take a release note section and render it into a blackfriday node in a tree.
func renderSectionAsNodes(section Section) []*blackfriday.Node {
	nodes := []*blackfriday.Node{}

	// render heading
	headingNode := blackfriday.NewNode(blackfriday.Heading)
	headingNode.Level = 2
	headingTextNode := blackfriday.NewNode(blackfriday.Text)
	headingTextNode.Literal = []byte(section.Heading)
	headingNode.AppendChild(headingTextNode)
	nodes = append(nodes, headingNode)

	// render preamble
	nodes = append(nodes, section.Preamble...)

	// render list
	listNode := blackfriday.NewNode(blackfriday.List)
	if section.Details.IsOrdered {
		listNode.ListFlags = blackfriday.ListTypeOrdered
	}
	seen := map[string]bool{}
	for _, node := range section.Details.Items {
		renderedNode := nodeAsString(node)
		_, hasSeen := seen[renderedNode]
		if !hasSeen {
			listNode.AppendChild(node)
			seen[renderedNode] = true
		}
	}
	nodes = append(nodes, listNode)

	return nodes
}

// nodeAsString will use the markdownfmt renderer to render the markdown tree into a string.
func nodeAsString(node *blackfriday.Node) string {
	buf := bytes.NewBufferString("")
	renderer := markdown.NewRenderer(nil)
	node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(buf, node, entering)
	})
	return buf.String()
}
