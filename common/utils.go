package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"os"
)

const (
	EnvServices      = "SERVICES"
	EnvAwsRegion     = "AWS_REGION" // reserved env
	EnvCustomGroups  = "CUSTOM_GROUPS"
	EnvSecretEnabled = "SECRET_ENABLED"
)

func GetSession() (*session.Session, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(os.Getenv(EnvAwsRegion)),
		},
	})

	if err != nil {
		return nil, fmt.Errorf("error occurred while trying to create a connection to aws: %s. Aborting", err.Error())
	}

	return sess, nil
}
