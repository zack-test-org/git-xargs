package main

import (
	"fmt"
)

type UnknownHeadingError struct {
	heading string
	body    string
}

func (err UnknownHeadingError) Error() string {
	return fmt.Sprintf("Unknown heading %s while parsing release note body: %s", err.heading, err.body)
}

type ReleaseNoteParsingError struct {
	body string
}

func (err ReleaseNoteParsingError) Error() string {
	return fmt.Sprintf("Error while parsing release note body %s", err.body)
}
