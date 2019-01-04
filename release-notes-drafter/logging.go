package main

import (
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/sirupsen/logrus"
)

// GetProjectLogger returns the logger object with global settings for the app.
func GetProjectLogger() *logrus.Entry {
	logger := logging.GetLogger("")
	return logger.WithField("name", "release-notes-drafter")
}
