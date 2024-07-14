package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"go.uber.org/zap"
	"os"
)

var sugLog *zap.SugaredLogger

// Ensure sugLog is safely initialized before use
func initLogger() {
	// Basic logger initialization, replace with your actual logger configuration
	logger, err := zap.NewProduction()
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1) // Or handle the error according to your application's requirements
	}
	sugLog = logger.Sugar()
}

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

func ValidateRequired() error {
	destinationArn := os.Getenv(envFirehoseArn)
	if destinationArn == emptyString {
		return fmt.Errorf("destination ARN must be set")
	}

	accountId := os.Getenv(envAccountId)
	if accountId == emptyString {
		return fmt.Errorf("account id must be set")
	}

	awsPartition := os.Getenv(envAwsPartition)
	if awsPartition == emptyString {
		return fmt.Errorf("aws partition must be set")
	}

	return nil
}
