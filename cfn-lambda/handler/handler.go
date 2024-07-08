package handler

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/logzio/firehose-logs/common"
	lp "github.com/logzio/firehose-logs/logger"
	"go.uber.org/zap"
)

const (
	secretEnabledKey   = "SecretEnabled"
	customLogGroupsKey = "CustomLogGroups"
	servicesKey        = "Services"
)

var sugLog *zap.SugaredLogger

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
	sugLog.Debug("Generated physicalId with value: ", physicalResourceId)
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
	oldServicesStr, err := extractConfigString(oldConfig, servicesKey)
	if err != nil {
		return err
	}
	newServicesStr, err := extractConfigString(newConfig, servicesKey)
	if err != nil {
		return err
	}
	oldCustomGroupsStr, err := extractConfigString(oldConfig, customLogGroupsKey)
	if err != nil {
		return err
	}
	oldSecretEnabledStr, err := extractConfigString(oldConfig, secretEnabledKey)
	if err != nil {
		return err
	}
	newCustomGroupsStr, err := extractConfigString(newConfig, customLogGroupsKey)
	if err != nil {
		return err
	}
	newSecretEnabledStr, err := extractConfigString(newConfig, secretEnabledKey)
	if err != nil {
		return err
	}

	oldCustomGroupsStr, err = common.GetCustomLogGroups(oldSecretEnabledStr, oldCustomGroupsStr)
	if err != nil {
		return err
	}
	newCustomGroupsStr, err = common.GetCustomLogGroups(newSecretEnabledStr, newCustomGroupsStr)
	if err != nil {
		return err
	}

	// Parse services and custom log groups
	oldServices := common.ParseServices(oldServicesStr)
	newServices := common.ParseServices(newServicesStr)
	oldCustomGroups := common.ParseServices(oldCustomGroupsStr)
	newCustomGroups := common.ParseServices(newCustomGroupsStr)

	// Find differences in services and custom log groups
	servicesToAdd, servicesToRemove := common.FindDifferences(oldServices, newServices)
	customGroupsToAdd, customGroupsToRemove := common.FindDifferences(oldCustomGroups, newCustomGroups)

	// Update subscription filters
	if err := common.UpdateSubscriptionFilters(sess, servicesToAdd, servicesToRemove, customGroupsToAdd, customGroupsToRemove); err != nil {
		sugLog.Errorf("Error updating subscription filters: %v", err)
		return err
	}

	return nil
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
		newDeleted, err := common.DeleteServices(sess, servicesToDelete)
		deleted = append(deleted, newDeleted...)
		if err != nil {
			sugLog.Error(err.Error())
		}
	}

	pathsToDelete := common.GetCustomPaths()
	if pathsToDelete != nil {
		newDeleted, err := common.DeleteCustom(sess, pathsToDelete)
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
		newAdded, err := common.AddServices(sess, servicesToAdd)
		added = append(added, newAdded...)
		if err != nil {
			sugLog.Error(err.Error())
		}
	}

	pathsToAdd := common.GetCustomPaths()
	if pathsToAdd != nil {
		newAdded, err := common.AddCustom(sess, pathsToAdd, added)
		added = append(added, newAdded...)
		if err != nil {
			sugLog.Error(err.Error())
		}
	}

	sugLog.Info("Following these log groups: ", added)

	return nil
}
