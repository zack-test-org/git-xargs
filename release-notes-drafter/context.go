package main

import (
	"os"
	"time"

	"github.com/sirupsen/logrus"
)

var (
	// GithubWebhookSecretkey is the secret key used to validate the incoming webhook requests so that we can make sure
	// it is coming from github. Only used for local or lambda modes.
	GithubWebhookSecretKey string
	// GithubApiKey is the personal access token to use to read pull request info and update release notes.
	GithubApiKey string
	// AwsRegion is the AWS region that should be used to access expected AWS resources.
	ProjectAwsRegion string
	// LockTableName is the name of the dynamodb table that should be used for the mutex locks
	ProjectLockTableName string
	// ProjectLockRetryTimeout is the timeout for attempting to acquire the lock before erroring out the request.
	ProjectLockRetryTimeout time.Duration
)

// GetSecret will lookup the secret in the following order, returning empty string if it is not defined in any of these:
// - AWS secret manager
// - Runtime environment variable
func GetSecret(logger *logrus.Entry, secretId string) string {
	maybeSecret, err := LookupSecret(secretId)
	if err != nil || maybeSecret == "" {
		logger.Warnf("Error looking up secret %s in AWS: %s", secretId, err)
		logger.Warn("Falling back to os environment")
		maybeSecret := os.Getenv(secretId)
		if maybeSecret != "" {
			logger.Infof("Loaded secret %s from OS environment", secretId)
		}
		return maybeSecret
	}
	logger.Infof("Loaded secret %s from AWS Secrets Manager", secretId)
	return maybeSecret
}

// SetContext will set the runtime context so that they can be accessed during the lifetime of the command.
func SetContext(logger *logrus.Entry, region string, lockTable string, lockTimeout time.Duration) {
	logger.Infof("Set runtime context: AWS Region = %s", region)
	ProjectAwsRegion = region

	logger.Infof("Set runtime context: Lock table name = %s", lockTable)
	ProjectLockTableName = lockTable

	logger.Infof("Set runtime context: Lock timeout = %s", lockTimeout)
	ProjectLockRetryTimeout = lockTimeout

	GithubWebhookSecretKey = GetSecret(logger, "GITHUB_WEBHOOK_SECRET")
	if GithubWebhookSecretKey == "" {
		logger.Warn("Could not find a value for Github webhook secret key in runtime environment")
	}
	GithubApiKey = GetSecret(logger, "GITHUB_TOKEN")
	if GithubApiKey == "" {
		logger.Warn("Could not find a value for Github API token in runtime environment")
	}
}
