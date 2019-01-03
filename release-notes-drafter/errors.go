package main

import (
	"fmt"
)

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

// ReleaseNoteParsingError is returned when the parser fails to parse the release note markdown into the ReleaseNote
// struct.
type ReleaseNoteParsingError struct {
	body string
}

func (err ReleaseNoteParsingError) Error() string {
	return fmt.Sprintf("Error while parsing release note body %s", err.body)
}
