package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"os"
)

func getLogGroups(services []string, logsClient *cloudwatchlogs.CloudWatchLogs) []string {
	if sugLog == nil {
		initLogger()
	}
	logGroupsToAdd := make([]string, 0)
	serviceToPrefix := GetServicesMap()
	for _, service := range services {
		if prefix, ok := serviceToPrefix[service]; ok {
			sugLog.Debug("Working on prefix: ", prefix)
			newLogGroups, err := logGroupsPagination(prefix, logsClient)
			if err != nil {
				sugLog.Errorf("Error while searching for log groups of %s: %s", service, err.Error())
			}

			logGroupsToAdd = append(logGroupsToAdd, newLogGroups...)
		} else {
			sugLog.Errorf("Service %s is not supported. Skipping.", service)
		}
	}

	return logGroupsToAdd
}

func logGroupsPagination(prefix string, logsClient *cloudwatchlogs.CloudWatchLogs) ([]string, error) {
	var nextToken *string
	logGroups := make([]string, 0)
	for {
		describeOutput, err := logsClient.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
			LogGroupNamePrefix: &prefix,
			NextToken:          nextToken,
		})

		if err != nil {
			return nil, err
		}
		if describeOutput != nil {
			nextToken = describeOutput.NextToken
			for _, logGroup := range describeOutput.LogGroups {
				// Prevent a situation where we put subscription filter on the trigger and shipper function
				if *logGroup.LogGroupName != LambdaPrefix+os.Getenv(EnvFunctionName) {
					logGroups = append(logGroups, *logGroup.LogGroupName)
				}
			}
		}

		if nextToken == nil {
			break
		}
	}

	return logGroups, nil
}

func AddServices(sess *session.Session, servicesToAdd []string) ([]string, error) {
	if sugLog == nil {
		initLogger()
	}
	logsClient := cloudwatchlogs.New(sess)
	logGroups := getLogGroups(servicesToAdd, logsClient)
	if len(logGroups) > 0 {
		sugLog.Debug("Detected the following services: ", logGroups)
		newAdded := PutSubscriptionFilter(logGroups, logsClient)
		return newAdded, nil
	} else {
		return nil, fmt.Errorf("Could not retrieve any log groups")
	}
}

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

func UpdateSubscriptionFilters(sess *session.Session, servicesToAdd, servicesToRemove, customGroupsToAdd, customGroupsToRemove []string) error {
	if sugLog == nil {
		initLogger()
	}
	// Add subscription filters for new services
	if len(servicesToAdd) > 0 {
		addedServices, err := AddServices(sess, servicesToAdd)
		if err != nil {
			sugLog.Errorf("Error adding subscriptions for services: %v", err)
			return err
		}
		sugLog.Infof("Added subscriptions for services: %v", addedServices)
	}

	// Add subscription filters for new custom log groups
	if len(customGroupsToAdd) > 0 {
		addedCustomGroups, err := AddCustom(sess, customGroupsToAdd, nil) // Assuming the third parameter is handled within the function
		if err != nil {
			sugLog.Errorf("Error adding subscriptions for custom log groups: %v", err)
			return err
		}
		sugLog.Infof("Added subscriptions for custom log groups: %v", addedCustomGroups)
	}

	// Remove subscription filters from services no longer needed
	if len(servicesToRemove) > 0 {
		removedServices, err := DeleteServices(sess, servicesToRemove)
		if err != nil {
			sugLog.Errorf("Error removing subscriptions for services: %v", err)
			return err
		}
		sugLog.Infof("Removed subscriptions for services: %v", removedServices)
	}

	// Remove subscription filters from custom log groups no longer needed
	if len(customGroupsToRemove) > 0 {
		removedCustomGroups, err := DeleteCustom(sess, customGroupsToRemove)
		if err != nil {
			sugLog.Errorf("Error removing subscriptions for custom log groups: %v", err)
			return err
		}
		sugLog.Infof("Removed subscriptions for custom log groups: %v", removedCustomGroups)
	}

	return nil
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

func DeleteServices(sess *session.Session, servicesToDelete []string) ([]string, error) {
	if sugLog == nil {
		initLogger()
	}
	logsClient := cloudwatchlogs.New(sess)
	logGroups := getLogGroups(servicesToDelete, logsClient)

	sugLog.Infow("Attempting to delete subscription filters",
		"servicesToDelete", servicesToDelete,
		"logGroups", logGroups)

	if len(logGroups) > 0 {
		newDeleted := DeleteSubscriptionFilter(logGroups, logsClient)
		sugLog.Infow("Deleted subscription filters",
			"deletedLogGroups", newDeleted)
		return newDeleted, nil
	} else {
		sugLog.Info("No log groups found for deletion")
		return nil, fmt.Errorf("Could not delete any log groups")
	}
}

func AddCustom(sess *session.Session, customGroup, added []string) ([]string, error) {
	logsClient := cloudwatchlogs.New(sess)
	toAdd := make([]string, 0)
	lambdaNameTrigger := LambdaPrefix + os.Getenv(EnvFunctionName)
	for _, customLogGroup := range customGroup {
		if !ListContains(customLogGroup, added) {
			// Prevent a situation where we put subscription filter on the trigger function
			if customLogGroup != lambdaNameTrigger {
				toAdd = append(toAdd, customLogGroup)
			}
		}
	}

	newAdded := PutSubscriptionFilter(toAdd, logsClient)

	return newAdded, nil
}

func DeleteCustom(sess *session.Session, customGroup []string) ([]string, error) {
	if sugLog == nil {
		initLogger()
	}
	logsClient := cloudwatchlogs.New(sess)

	newDeleted := DeleteSubscriptionFilter(customGroup, logsClient)

	// Log the outcome of the deletion attempts
	sugLog.Infow("Deleted custom subscription filters",
		"deletedCustomLogGroups", newDeleted)

	return newDeleted, nil
}
