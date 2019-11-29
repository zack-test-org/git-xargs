package aws

import (
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/gruntwork-io/gruntwork-cli/errors"
)

const DefaultRegion = "us-east-2"

func NewAuthenticatedSession() (*session.Session, error) {
	sess, err := session.NewSession(aws.NewConfig().WithRegion(DefaultRegion))
	if err != nil {
		return nil, errors.WithStackTrace(err)
	}

	if _, err = sess.Config.Credentials.Get(); err != nil {
		return nil, errors.WithStackTrace(err)
	}

	return sess, nil
}
