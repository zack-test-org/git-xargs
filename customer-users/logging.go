package main

import (
	"github.com/gruntwork-io/gruntwork-cli/logging"
	"github.com/sirupsen/logrus"
)

func GetProjectLogger() *logrus.Entry {
	logger := logging.GetLogger("")
	return logger.WithField("name", "customer-users")
}
