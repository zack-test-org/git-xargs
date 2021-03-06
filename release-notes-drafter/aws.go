package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/dynamodb"
	"github.com/aws/aws-sdk-go/service/secretsmanager"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

// NewAuthenticatedSession gets an AWS Session, checking that the user has credentials properly configured in their environment.
func NewAuthenticatedSession() (*session.Session, error) {
	sess, err := session.NewSession(aws.NewConfig().WithRegion(ProjectAwsRegion))
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if _, err = sess.Config.Credentials.Get(); err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return sess, nil
}

// LookupSecret will lookup the value of the requested secret.
func LookupSecret(secretName string) (string, error) {
	sess, err := NewAuthenticatedSession()
	if err != nil {
		return "", err
	}
	secmgrSvc := secretsmanager.New(sess)
	secretId := fmt.Sprintf("release-notes-drafter/%s", secretName)
	resp, err := secmgrSvc.GetSecretValue(&secretsmanager.GetSecretValueInput{SecretId: aws.String(secretId)})
	if err != nil {
		return "", errors.WithStackTrace(err)
	}

	// Depending on whether the secret is a string or binary, one of these fields will be populated.
	if resp.SecretString != nil {
		return aws.StringValue(resp.SecretString), nil
	}
	return decodeSecretBinary(resp.SecretBinary)
}

// decodeSecretBinary will decode the raw base64 encoded binary data from the Secrets Manager API into a string.
// This is the same code that is included in the Secrets Manager UI in the AWS console.
func decodeSecretBinary(secretBinary []byte) (string, error) {
	decodedBinarySecretBytes := make([]byte, base64.StdEncoding.DecodedLen(len(secretBinary)))
	numbytes, err := base64.StdEncoding.Decode(decodedBinarySecretBytes, secretBinary)
	if err != nil {
		return "", errors.WithStackTrace(err)
	}
	decodedBinarySecret := string(decodedBinarySecretBytes[:numbytes])
	return decodedBinarySecret, nil
}

// NewDynamoDb returns an authenticated client object for accessing DynamoDb
func NewDynamoDb() (*dynamodb.DynamoDB, error) {
	sess, err := NewAuthenticatedSession()
	if err != nil {
		return nil, err
	}
	dynamodbSvc := dynamodb.New(sess)
	return dynamodbSvc, nil
}

// AcquireLock will attempt to acquire the lock defined by the provided lock string in the configured lock table for the
// configured region.
func AcquireLock(lockString string) error {
	logger := GetProjectLogger()
	logger.Infof(
		"Attempting to acquire lock %s in table %s in region %s",
		lockString,
		ProjectLockTableName,
		ProjectAwsRegion,
	)

	dynamodbSvc, err := NewDynamoDb()
	if err != nil {
		logger.Errorf("Error authenticating to AWS: %s", err)
		return err
	}

	putParams := &dynamodb.PutItemInput{
		Item: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(lockString)},
		},
		TableName:           aws.String(ProjectLockTableName),
		ConditionExpression: aws.String("attribute_not_exists(LockID)"),
	}
	_, err = dynamodbSvc.PutItem(putParams)
	if err != nil {
		logger.Warnf(
			"Error acquiring lock %s in table %s in region %s (already locked?): %s",
			lockString,
			ProjectLockTableName,
			ProjectAwsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	logger.Info("Acquired lock")
	return nil
}

// BlockingAcquireLock will attempt to acquire the lock defined by the provided lock string in the configured lock table
// for the configured region. This will retry on failure, until reaching timeout.
func BlockingAcquireLock(lockString string) error {
	logger := GetProjectLogger()
	logger.Infof(
		"Attempting to acquire lock %s in table %s in region %s, retrying on failure for up to %s",
		lockString,
		ProjectLockTableName,
		ProjectAwsRegion,
		ProjectLockRetryTimeout,
	)

	// Timeout logic inspired by terratest
	// See: https://github.com/gruntwork-io/terratest/blob/master/modules/retry/retry.go
	ctx, cancel := context.WithTimeout(context.Background(), ProjectLockRetryTimeout)
	defer cancel()

	doneChannel := make(chan bool, 1)

	go func() {
		for AcquireLock(lockString) != nil {
			logger.Warnf("Failed to acquire lock %s. Retrying in 5 seconds...", lockString)
			time.Sleep(5 * time.Second)
		}
		doneChannel <- true
	}()
	select {
	case <-doneChannel:
		logger.Infof("Successfully acquired lock %s", lockString)
		return nil
	case <-ctx.Done():
		logger.Errorf("Timed out attempting to acquire lock %s", lockString)
		return LockTimeoutExceeded{LockTable: ProjectLockTableName, LockString: lockString, Timeout: ProjectLockRetryTimeout}
	}
}

// ReleaseLock will attempt to release the lock defined by the provided lock string in the configured lock table for the
// configured region.
func ReleaseLock(lockString string) error {
	logger := GetProjectLogger()
	logger.Infof(
		"Attempting to release lock %s in table %s in region %s",
		lockString,
		ProjectLockTableName,
		ProjectAwsRegion,
	)

	dynamodbSvc, err := NewDynamoDb()
	if err != nil {
		logger.Errorf("Error authenticating to AWS: %s", err)
		return err
	}

	params := &dynamodb.DeleteItemInput{
		Key: map[string]*dynamodb.AttributeValue{
			"LockID": {S: aws.String(lockString)},
		},
		TableName: aws.String(ProjectLockTableName),
	}
	_, err = dynamodbSvc.DeleteItem(params)

	if err != nil {
		logger.Errorf(
			"Error releasing lock %s in table %s in region %s: %s",
			lockString,
			ProjectLockTableName,
			ProjectAwsRegion,
			err,
		)
		return errors.WithStackTrace(err)
	}
	logger.Info("Released lock")
	return nil
}
