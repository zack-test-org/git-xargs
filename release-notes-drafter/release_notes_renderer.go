package main

import (
	"bytes"

	"github.com/jsternberg/markdownfmt/markdown"
	blackfriday "gopkg.in/russross/blackfriday.v2"
)

func RenderReleaseNote(releaseNote ReleaseNote) string {
	rootNode := blackfriday.NewNode(blackfriday.Document)

	for _, node := range RenderSectionAsNodes(releaseNote.ModulesAffected) {
		rootNode.AppendChild(node)
	}
	for _, node := range RenderSectionAsNodes(releaseNote.Description) {
		rootNode.AppendChild(node)
	}
	for _, node := range RenderSectionAsNodes(releaseNote.RelatedLinks) {
		rootNode.AppendChild(node)
	}

	return nodeAsString(rootNode)
}

func RenderSectionAsNodes(section Section) []*blackfriday.Node {
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

func nodeAsString(node *blackfriday.Node) string {
	buf := bytes.NewBufferString("")
	renderer := markdown.NewRenderer(nil)
	node.Walk(func(node *blackfriday.Node, entering bool) blackfriday.WalkStatus {
		return renderer.RenderNode(buf, node, entering)
	})
	return buf.String()
}
