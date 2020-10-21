package cmd

import (
	"strconv"

	"github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
)

var (
	WorkflowsNode *yaml.Node
)

// Recursively search for a node with a given Value string
func iterateNode(node *yaml.Node, identifier string) *yaml.Node {
	returnNode := false
	for _, n := range node.Content {
		if n.Value == identifier {
			returnNode = true
			continue
		}
		if returnNode {
			return n
		}
		if len(n.Content) > 0 {
			a_node := iterateNode(n, identifier)
			if a_node != nil {
				return a_node
			}
		}
	}
	return nil
}

// Update any existing context nodes with a test string to prove out this approach, for the time being
// TODO:
// Ultimately, this method should intelligently handle all applicable cases:
// - A context array does not exist and needs to be created and populated
// - A context array exists but is missing the correct context
// - A context array exists and already has the correct context
//
func updateContextNodes(node *yaml.Node) {
	for i, n := range node.Content {
		// If node is a context node, then its immediate neighbor contains the value
		// TODO: add out of bounds guard
		if n.Value == "context" {
			valueNode := node.Content[i+1]
			valueNode.Content[0].Value = "I AM INTENTIONALLY OVERWRITING THIS VALUE AS A TEST"
		}
		if len(n.Content) > 0 {
			updateContextNodes(n)
		}
	}
}

// Entrypoint for YAML processing
// Accepts a byte slice containing YAML, marshals it into a *yaml.Node tree, and traverses the tree to modify the Workflows -> Jobs -> Context nodes in place
func updateYamlDocument(b []byte) []byte {
	// yaml node tree that will contain the will represent the entirety of a given YAML document in *yaml.Nodes
	// See: https://ubuntu.com/blog/api-v3-of-the-yaml-package-for-go-is-available
	var c yaml.Node
	unmarshalErr := yaml.Unmarshal([]byte(b), &c)
	if unmarshalErr != nil {
		log.WithFields(logrus.Fields{
			"Error": unmarshalErr,
		}).Debug("Error unmarshaling YAML document to *yaml.Node tree")
	}

	// Find workflows node in the parent "document" / root node. The root of the *yamlNode tree is of "Kind" "DocumentNode"
	// See: https://pkg.go.dev/gopkg.in/yaml.v3#Kind
	for i := 0; i < len(c.Content[0].Content); i++ {
		nextNode := c.Content[0].Content[i+1]
		if c.Content[0].Content[i].Value == "workflows" && nextNode.Kind == yaml.MappingNode {
			WorkflowsNode = nextNode
			break
		}
	}

	// Find the version node from the Workflows map ensure it is at least 2 (support for contexts exists in versions 2.0+)
	vn := iterateNode(WorkflowsNode, "version")

	if vn != nil {
		parsedInt, parseErr := strconv.Atoi(vn.Value)
		if parseErr != nil {
			log.WithFields(logrus.Fields{
				"Error": parseErr,
			}).Debug("Error parsing Workflows syntax version")
		}
		// Found workflows version
		if parsedInt < 2 {
			log.WithFields(logrus.Fields{
				"Error": "Wokflows syntax version is less than 2.0, which contains support for contexts",
			}).Debug("Worfklows syntax version too low")
		}
	}

	// Find the Workflows -> Jobs node
	jobsNode := iterateNode(WorkflowsNode, "jobs")

	// Within that jobs node, look for and update the context nodes
	updateContextNodes(jobsNode)

	updatedYamlBytes, marshalErr := yaml.Marshal(&c)

	if marshalErr != nil {
		log.WithFields(logrus.Fields{
			"Error": marshalErr,
		}).Debug("Error marshaling updated YAML")
	}

	return updatedYamlBytes

}
