package main

import (
	"github.com/gruntwork-io/gruntwork-cli/errors"
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/sirupsen/logrus"
)

// GetProjectLogger returns the logger object with global settings for the app.
func GetProjectLogger() *logrus.Entry {
	logger := logging.GetLogger("")
	return logger.WithField("name", "release-notes-drafter")
}

// LogIfError is a convenience function that will log the error and return the error wrapped in StackTrace if it is an
// error, otherwise return nil.
func LogIfError(logger *logrus.Entry, logMessage string, err error) error {
	if err != nil {
		logger.Errorf("%s: %s", logMessage, err)
		return errors.WithStackTrace(err)
	}
	return nil
}
