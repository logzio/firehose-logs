package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"go.uber.org/zap"
	"os"
)

var sugLog *zap.SugaredLogger

func PutSubscriptionFilter(logGroups []string, logsClient *cloudwatchlogs.CloudWatchLogs) []string {
	// Early return if logsClient is nil to avoid panic
	if logsClient == nil {
		fmt.Println("CloudWatch Logs client is nil")
		return nil
	}

	// Initialize logger if it's nil
	if sugLog == nil {
		initLogger()
	}

	destinationArn := os.Getenv(envFirehoseArn)
	roleArn := os.Getenv(envPutSubscriptionFilterRole)
	filterPattern := ""
	filterName := subscriptionFilterName
	added := make([]string, 0)
	for _, logGroup := range logGroups {
		_, err := logsClient.PutSubscriptionFilter(&cloudwatchlogs.PutSubscriptionFilterInput{
			DestinationArn: &destinationArn,
			FilterName:     &filterName,
			LogGroupName:   &logGroup,
			FilterPattern:  &filterPattern,
			RoleArn:        &roleArn,
		})

		if err != nil {
			sugLog.Error("Error while trying to add subscription filter for ", logGroup, ": ", err.Error())
			continue
		}

		added = append(added, logGroup)
	}

	return added
}

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

func DeleteSubscriptionFilter(logGroups []string, logsClient *cloudwatchlogs.CloudWatchLogs) []string {
	// Early return if logsClient is nil to avoid panic
	if logsClient == nil {
		fmt.Println("CloudWatch Logs client is nil")
		return nil
	}

	// Initialize logger if it's nil
	if sugLog == nil {
		initLogger()
	}

	filterName := subscriptionFilterName
	deleted := make([]string, 0)
	for _, logGroup := range logGroups {
		_, err := logsClient.DeleteSubscriptionFilter(&cloudwatchlogs.DeleteSubscriptionFilterInput{
			FilterName:   &filterName,
			LogGroupName: &logGroup,
		})

		if err != nil {
			sugLog.Error("Error while trying to delete subscription filter for ", logGroup, ": ", err.Error())
			continue
		}

		deleted = append(deleted, logGroup)
		sugLog.Info("Detected the following services for deletion2: ", deleted)

	}

	return deleted
}

func GetSession() (*session.Session, error) {
	sess, err := session.NewSessionWithOptions(session.Options{
		Config: aws.Config{
			Region: aws.String(os.Getenv(envAwsRegion)),
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
