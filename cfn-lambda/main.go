package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/logzio/firehose-logs/common"
	lp "github.com/logzio/firehose-logs/logger"
	"go.uber.org/zap"
	"os"
)

var sugLog *zap.SugaredLogger

func main() {
	lambda.Start(cfn.LambdaWrap(HandleRequest))
}

func HandleRequest(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	logger := lp.GetLogger()
	defer logger.Sync()
	sugLog = logger.Sugar()
	sugLog.Info("Starting handling event...")
	sugLog.Debug("Handling event: ", event)
	err = common.ValidateRequired()
	if err != nil {
		sugLog.Debug("Lambda finished with error")
		return "", nil, err
	}

	// CloudFormation custom resource handling logic
	switch event.RequestType {
	case "Create":
		sugLog.Debug("Detected CloudFormation Create event")
		return customResourceRun(ctx, event)
	case "Update":
		sugLog.Debug("Detected CloudFormation Update event")
		return customResourceRunUpdate(ctx, event)
	case "Delete":
		sugLog.Debug("Detected CloudFormation Delete event")
		return customResourceRunDelete(ctx, event)
	default:
		sugLog.Debug("Detected unsupported request type")
		return customResourceRunDoNothing(ctx, event)
	}
}

func generatePhysicalResourceId(event cfn.Event) string {
	// Concatenate StackId and LogicalResourceId to form a unique PhysicalResourceId
	physicalResourceId := fmt.Sprintf("%s-%s", event.StackID, event.LogicalResourceID)
	return physicalResourceId
}

func customResourceRunUpdate(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	oldConfig := event.OldResourceProperties
	newConfig := event.ResourceProperties

	physicalResourceID = generatePhysicalResourceId(event)

	err = updateConfiguration(ctx, oldConfig, newConfig)
	if err != nil {
		sugLog.Error("Error during update: ", err)
		return physicalResourceID, nil, err
	}

	// Populate your data map as needed for the update
	data = make(map[string]interface{})
	// Populate data as needed

	return physicalResourceID, data, nil
}

func updateConfiguration(ctx context.Context, oldConfig, newConfig map[string]interface{}) error {
	sess, err := common.GetSession()
	if err != nil {
		sugLog.Error("Error while creating session: ", err.Error())
		return err
	}

	sugLog.Info("Extracting configuration strings...")

	// Helper function to extract and validate configuration strings
	extractConfigString := func(config map[string]interface{}, key string) (string, error) {
		value, exists := config[key]
		if !exists {
			return "", nil
		}
		strValue, ok := value.(string)
		if !ok {
			sugLog.Errorf("Invalid type for %s; expected string", key)
			return "", fmt.Errorf("invalid configuration type for %s", key)
		}
		return strValue, nil
	}

	// Extract and validate services and custom log group strings from the configurations
	oldServicesStr, err := extractConfigString(oldConfig, "Services")
	newServicesStr, err := extractConfigString(newConfig, "Services")
	oldCustomGroupsStr, err := extractConfigString(oldConfig, "CustomLogGroups")
	newCustomGroupsStr, err := extractConfigString(newConfig, "CustomLogGroups")
	if err != nil {
		return err
	}

	// Parse services and custom log groups
	oldServices := common.ParseServices(oldServicesStr)
	newServices := common.ParseServices(newServicesStr)
	oldCustomGroups := common.ParseServices(oldCustomGroupsStr)
	newCustomGroups := common.ParseServices(newCustomGroupsStr)

	// Find differences in services and custom log groups
	servicesToAdd, servicesToRemove := findDifferences(oldServices, newServices)
	customGroupsToAdd, customGroupsToRemove := findDifferences(oldCustomGroups, newCustomGroups)

	// Update subscription filters
	if err := updateSubscriptionFilters(sess, servicesToAdd, servicesToRemove, customGroupsToAdd, customGroupsToRemove); err != nil {
		sugLog.Errorf("Error updating subscription filters: %v", err)
		return err
	}

	return nil
}

func updateSubscriptionFilters(sess *session.Session, servicesToAdd, servicesToRemove, customGroupsToAdd, customGroupsToRemove []string) error {
	// Add subscription filters for new services
	if len(servicesToAdd) > 0 {
		addedServices, err := addServices(sess, servicesToAdd)
		if err != nil {
			sugLog.Errorf("Error adding subscriptions for services: %v", err)
			return err
		}
		sugLog.Infof("Added subscriptions for services: %v", addedServices)
	}

	// Add subscription filters for new custom log groups
	if len(customGroupsToAdd) > 0 {
		addedCustomGroups, err := addCustom(sess, customGroupsToAdd, nil) // Assuming the third parameter is handled within the function
		if err != nil {
			sugLog.Errorf("Error adding subscriptions for custom log groups: %v", err)
			return err
		}
		sugLog.Infof("Added subscriptions for custom log groups: %v", addedCustomGroups)
	}

	// Remove subscription filters from services no longer needed
	if len(servicesToRemove) > 0 {
		removedServices, err := deleteServices(sess, servicesToRemove)
		if err != nil {
			sugLog.Errorf("Error removing subscriptions for services: %v", err)
			return err
		}
		sugLog.Infof("Removed subscriptions for services: %v", removedServices)
	}

	// Remove subscription filters from custom log groups no longer needed
	if len(customGroupsToRemove) > 0 {
		removedCustomGroups, err := deleteCustom(sess, customGroupsToRemove)
		if err != nil {
			sugLog.Errorf("Error removing subscriptions for custom log groups: %v", err)
			return err
		}
		sugLog.Infof("Removed subscriptions for custom log groups: %v", removedCustomGroups)
	}

	return nil
}

// findDifferences finds elements in 'new' that are not in 'old', and vice versa.
func findDifferences(old, new []string) (toAdd, toRemove []string) {
	oldSet := make(map[string]struct{})
	newSet := make(map[string]struct{})

	// Populate 'oldSet' with elements from the 'old' slice.
	for _, item := range old {
		oldSet[item] = struct{}{}
	}

	for _, item := range new {
		newSet[item] = struct{}{}
	}

	// Find elements in 'new' that are not in 'old' and add them to 'toAdd'.
	for item := range newSet {
		_, exists := oldSet[item] // Check if 'item' exists in 'oldSet'
		if !exists {
			toAdd = append(toAdd, item)
		}
	}

	for item := range oldSet {
		_, exists := newSet[item]
		if !exists {
			toRemove = append(toRemove, item)
		}
	}

	return toAdd, toRemove
}

// Wrapper for first invocation from cloud formation custom resource
func customResourceRun(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	physicalResourceID = generatePhysicalResourceId(event)

	err = handleFirstInvocation()
	if err != nil {
		sugLog.Error("Error while handling first invocation: ", err.Error())
		return physicalResourceID, nil, err
	}

	// Populate your data map as needed for the update
	data = make(map[string]interface{})
	// Populate data as needed

	return physicalResourceID, data, nil
}

// Wrapper for invocation from cloudformation custom resource - for read, update
func customResourceRunDoNothing(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	return
}

// Wrapper for invocation from cloudformation custom resource - delete
func customResourceRunDelete(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	sess, err := common.GetSession()
	if err != nil {
		sugLog.Error("Error while creating session: ", err.Error())
	}

	deleted := make([]string, 0)
	servicesToDelete := common.GetServices()
	if servicesToDelete != nil {
		newDeleted, err := deleteServices(sess, servicesToDelete)
		deleted = append(deleted, newDeleted...)
		if err != nil {
			sugLog.Error(err.Error())
		}
	}

	pathsToDelete := common.GetCustomPaths()
	if pathsToDelete != nil {
		newDeleted, err := deleteCustom(sess, pathsToDelete)
		deleted = append(deleted, newDeleted...)
		if err != nil {
			sugLog.Error(err.Error())
		}
	}

	sugLog.Info("Deleted subscription filters for the following log groups: ", deleted)

	physicalResourceID = generatePhysicalResourceId(event)
	// Populate your data map as needed for the update
	data = make(map[string]interface{})
	// Populate data as needed

	return physicalResourceID, data, nil
}

func handleFirstInvocation() error {
	sess, err := common.GetSession()
	if err != nil {
		return err
	}

	added := make([]string, 0)
	servicesToAdd := common.GetServices()
	if servicesToAdd != nil {
		newAdded, err := addServices(sess, servicesToAdd)
		added = append(added, newAdded...)
		if err != nil {
			sugLog.Error(err.Error())
		}
	}

	pathsToAdd := common.GetCustomPaths()
	if pathsToAdd != nil {
		newAdded, err := addCustom(sess, pathsToAdd, added)
		added = append(added, newAdded...)
		if err != nil {
			sugLog.Error(err.Error())
		}
	}

	sugLog.Info("Following these log groups: ", added)

	return nil
}

func addCustom(sess *session.Session, customGroup, added []string) ([]string, error) {
	logsClient := cloudwatchlogs.New(sess)
	toAdd := make([]string, 0)
	lambdaNameTrigger := common.LambdaPrefix + os.Getenv(common.EnvFunctionName)
	for _, customLogGroup := range customGroup {
		if !common.ListContains(customLogGroup, added) {
			// Prevent a situation where we put subscription filter on the trigger function
			if customLogGroup != lambdaNameTrigger {
				toAdd = append(toAdd, customLogGroup)
			}
		}
	}

	newAdded := common.PutSubscriptionFilter(toAdd, logsClient)

	return newAdded, nil
}

func addServices(sess *session.Session, servicesToAdd []string) ([]string, error) {
	logsClient := cloudwatchlogs.New(sess)
	logGroups := getLogGroups(servicesToAdd, logsClient)
	if len(logGroups) > 0 {
		sugLog.Debug("Detected the following services: ", logGroups)
		newAdded := common.PutSubscriptionFilter(logGroups, logsClient)
		return newAdded, nil
	} else {
		return nil, fmt.Errorf("Could not retrieve any log groups")
	}
}

func getLogGroups(services []string, logsClient *cloudwatchlogs.CloudWatchLogs) []string {
	logGroupsToAdd := make([]string, 0)
	serviceToPrefix := common.GetServicesMap()
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
				if *logGroup.LogGroupName != common.LambdaPrefix+os.Getenv(common.EnvFunctionName) {
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

func deleteServices(sess *session.Session, servicesToDelete []string) ([]string, error) {
	logsClient := cloudwatchlogs.New(sess)
	logGroups := getLogGroups(servicesToDelete, logsClient)

	sugLog.Infow("Attempting to delete subscription filters",
		"servicesToDelete", servicesToDelete,
		"logGroups", logGroups)

	if len(logGroups) > 0 {
		newDeleted := common.DeleteSubscriptionFilter(logGroups, logsClient)
		sugLog.Infow("Deleted subscription filters",
			"deletedLogGroups", newDeleted)
		return newDeleted, nil
	} else {
		sugLog.Info("No log groups found for deletion")
		return nil, fmt.Errorf("Could not delete any log groups")
	}
}

func deleteCustom(sess *session.Session, customGroup []string) ([]string, error) {
	logsClient := cloudwatchlogs.New(sess)

	newDeleted := common.DeleteSubscriptionFilter(customGroup, logsClient)

	// Log the outcome of the deletion attempts
	sugLog.Infow("Deleted custom subscription filters",
		"deletedCustomLogGroups", newDeleted)

	return newDeleted, nil
}
