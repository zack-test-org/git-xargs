package main

import (
	"fmt"
	"time"
)

// LockTimeoutExceeded is returned when we timeout trying to acquire the lock.
type LockTimeoutExceeded struct {
	LockTable  string
	LockString string
	Timeout    time.Duration
}

func (err LockTimeoutExceeded) Error() string {
	return fmt.Sprintf("Timedout trying to acquire lock %s in table %s (timeout was %s)", err.LockString, err.LockTable, err.Timeout)
}

// MissingRequiredParameter is returned when a required parameter is missing.
type MissingRequiredParameter struct {
	ParamName string
}

func (err MissingRequiredParameter) Error() string {
	return fmt.Sprintf("Missing required parameter %s", err.ParamName)
}

// IncorrectHandlerError is returned when an incorrect handler is invoked for the pull request event.
type IncorrectHandlerError struct {
	EventAction string
}

func (err IncorrectHandlerError) Error() string {
	return fmt.Sprintf("Incorrect handler invoked for pull request event %s", err.EventAction)
}

// MissingMarkerError is returned when a marker for the release notes is missing.
type MissingMarkerError struct {
	Marker string
	Body   string
}

func (err MissingMarkerError) Error() string {
	return fmt.Sprintf("Release note is missing marker \"%s\" in body \"%s\"", err.Marker, err.Body)
}
