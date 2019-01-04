package main

import (
	"fmt"
)

// MissingRequiredParameter is returned when a required parameter is missing.
type MissingRequiredParameter struct {
	paramName string
}

func (err MissingRequiredParameter) Error() string {
	return fmt.Sprintf("Missing required parameter %s", err.paramName)
}

// IncorrectParserError is returned when an incorrect parser is invoked for the markdown node.
type IncorrectParserError struct {
	nodeInfo string
}

func (err IncorrectParserError) Error() string {
	return fmt.Sprintf("Incorrect parser invoked for markdown node %s", err.nodeInfo)
}

// IncorrectHandlerError is returned when an incorrect handler is invoked for the pull request event.
type IncorrectHandlerError struct {
	eventAction string
}

func (err IncorrectHandlerError) Error() string {
	return fmt.Sprintf("Incorrect handler invoked for pull request event %s", err.eventAction)
}

// UnknownHeadingError is returned when the parser encounters a markdown heading that it does not expect or understand
type UnknownHeadingError struct {
	heading string
	body    string
}

func (err UnknownHeadingError) Error() string {
	return fmt.Sprintf("Unknown heading %s while parsing release note body: %s", err.heading, err.body)
}

// MissingMarkerError is returned when a marker for the release notes is missing.
type MissingMarkerError struct {
	marker string
	body   string
}

func (err MissingMarkerError) Error() string {
	return fmt.Sprintf("Release note is missing marker \"%s\" in body \"%s\"", err.marker, err.body)
}
