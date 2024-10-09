package handler

import (
	"context"
	"fmt"
	"github.com/logzio/firehose-logs/common"
	"github.com/logzio/firehose-logs/logger"
	"go.uber.org/zap"
	"os"
	"strings"
)

var sugLog *zap.SugaredLogger

func HandleRequest(ctx context.Context, event map[string]interface{}) (string, error) {
	sugLog = logger.GetSugaredLogger()

	err := validateRequired()
	if err != nil {
		return "Lambda finished with error", err
	}

	sugLog.Info("Starting handling event...")
	sugLog.Debug("Handling event: ", event)

	detail, ok := event["detail"].(map[string]interface{})
	if !ok {
		sugLog.Error("`detail` is not of type map[string]interface{} or missing from the event.")
	}

	eventName, ok := detail["eventName"].(string)
	if !ok {
		sugLog.Error("`eventName` is not of type string or missing from the event.")
	}

	requestParameters, ok := detail["requestParameters"].(map[string]interface{})
	if !ok {
		sugLog.Error("`requestParameters` is not of type map[string]interface{} or missing from the event.")
	}

	switch eventName {
	case "CreateLogGroup":
		sugLog.Debug("Detected EventBridge CreateLogGroup event")

		logGroup, ok := requestParameters["logGroupName"].(string)
		if !ok {
			sugLog.Error("`logGroupName` is not of type string or missing from EventBridge event")
			return "", fmt.Errorf("`logGroupName` is not of type string or missing from EventBridge event")
		}
		handleNewLogGroupEvent(ctx, logGroup)

	case "PutSecretValue":
		sugLog.Debug("Detected EventBridge PutSecretValue event")

		secretId, ok := requestParameters["secretId"].(string)
		if !ok {
			sugLog.Error("`secretId` is not of type string or missing from EventBridge event")
			return "", fmt.Errorf("`secretId` is not of type string or missing from EventBridge event")
		}
		err := handleSecretChangedEvent(ctx, secretId)
		if err != nil {
			return "", err
		}

	case "SubscriptionFilterEvent":
		sugLog.Debug("Detected SubscriptionFilterEvent event")

		var reqParams common.RequestParameters
		reqParams, err = common.ConvertToRequestParameters(requestParameters)
		if err != nil {
			sugLog.Error("Error converting request parameters: ", err.Error())
			return "", err
		}

		actionType := reqParams.Action
		switch actionType {
		case common.AddSF:
			sugLog.Debug("Detected Add Subscription Filter event")
			handleCreateEvent(ctx, reqParams)
		case common.UpdateSF:
			sugLog.Debug("Detected Update Subscription Filter event")
			handleUpdateEvent(ctx, reqParams)
		case common.DeleteSF:
			sugLog.Debug("Detected Delete Subscription Filter event")
			return handleDeleteEvent(ctx, reqParams)
		default:
			sugLog.Debug("Detected unsupported Subscription Filter event")
			return "", fmt.Errorf("unsupported Subscription Filter event")
		}

	default:
		sugLog.Debug("Detected unsupported event")
		return "", fmt.Errorf("unsupported event")
	}

	return fmt.Sprintf("%s event handled successfully", eventName), nil
}

func handleNewLogGroupEvent(ctx context.Context, newLogGroup string) {
	// Prevent a situation where we put subscription filter on the trigger function
	if newLogGroup == lambdaPrefix+os.Getenv(envFunctionName) {
		return
	}

	// Check if the log group is of a monitored service
	currMonitoredServices := getServices()
	var added []string
	if currMonitoredServices != nil {
		serviceToPrefix := getServicesMap()

		cwClient, err := getCloudWatchLogsClient()
		if err != nil {
			sugLog.Error("Failed to get cloudwatch logs client")
		}

		for _, service := range currMonitoredServices {
			if prefix, ok := serviceToPrefix[service]; ok {
				if strings.Contains(newLogGroup, prefix) {
					added, _ = cwClient.addSubscriptionFilter([]string{newLogGroup})
					if len(added) > 0 {
						sugLog.Info("Added subscription filter to log group: ", newLogGroup)
						return
					}
				}
			}
		}
	}

	// Check if the log group is of a monitored custom prefix
	currCustomGroupsPrefixes := getCustomGroupsPrefixes()
	if len(currCustomGroupsPrefixes) > 0 {
		cwClient, err := getCloudWatchLogsClient()
		if err != nil {
			sugLog.Error("Failed to get cloudwatch logs client")
		}

		for _, prefix := range currCustomGroupsPrefixes {
			if strings.Contains(newLogGroup, prefix) {
				added, _ = cwClient.addSubscriptionFilter([]string{newLogGroup})
				if len(added) > 0 {
					sugLog.Info("Added subscription filter to log group: ", newLogGroup)
					return
				}
			}
		}
	}
}

func handleSecretChangedEvent(ctx context.Context, secretId string) error {
	secretName := os.Getenv(common.EnvCustomGroups)

	// make sure that the secret which changed is the relevant secret
	if strings.Contains(secretId, secretName) {
		err := updateSecretCustomLogGroups(ctx, secretId)
		if err != nil {
			sugLog.Error("Error while updating secret custom log groups: ", err.Error())
			return err
		}
	} else {
		sugLog.Debug("The EventBridge event secretId is not the secret that has custom log groups in it. Skipping it.")
	}
	return nil
}

func handleCreateEvent(ctx context.Context, event common.RequestParameters) {
	cwClient, err := getCloudWatchLogsClient()
	if err != nil {
		sugLog.Error("Failed to get cloudwatch logs client")
		return
	}

	servicesToMonitor := convertStrToArr(event.NewServices)
	logGroupsToMonitor := getServicesLogGroups(servicesToMonitor, cwClient)

	customLogGroupsToMonitor, err := getCustomLogGroups(event.NewIsSecret, event.NewCustom)
	if err != nil {
		sugLog.Error("Error while getting custom log groups: ", err.Error())
	}
	logGroupsToMonitor = append(logGroupsToMonitor, customLogGroupsToMonitor...)

	added, _ := cwClient.addSubscriptionFilter(logGroupsToMonitor)

	sugLog.Info("Added subscription filters for the following log groups: ", added)
}

func handleUpdateEvent(ctx context.Context, event common.RequestParameters) {
	cwClient, err := getCloudWatchLogsClient()
	if err != nil {
		sugLog.Error("Failed to get cloudwatch logs client")
	}

	oldServices := convertStrToArr(event.OldServices)
	newServices := convertStrToArr(event.NewServices)

	oldCustomGroups, err := getCustomLogGroups(event.OldIsSecret, event.OldCustom)
	if err != nil {
		sugLog.Error("Error while getting old custom log groups: ", err.Error())
	}
	newCustomGroups, err := getCustomLogGroups(event.NewIsSecret, event.NewCustom)
	if err != nil {
		sugLog.Error("Error while getting new custom log groups: ", err.Error())
	}

	servicesToAdd, servicesToRemove := findDifferences(oldServices, newServices)
	customGroupsToAdd, customGroupsToRemove := findDifferences(oldCustomGroups, newCustomGroups)

	err = cwClient.updateSubscriptionFilters(servicesToAdd, servicesToRemove, customGroupsToAdd, customGroupsToRemove)
	if err != nil {
		sugLog.Error("Error while updating subscription filters: ", err.Error())
	}
}

func handleDeleteEvent(ctx context.Context, event common.RequestParameters) (string, error) {
	cwClient, err := getCloudWatchLogsClient()
	if err != nil {
		sugLog.Error("Failed to get cloudwatch logs client")
		return "", err
	}

	servicesToUnMonitor := convertStrToArr(event.NewServices)
	logGroupsToUnMonitor := getServicesLogGroups(servicesToUnMonitor, cwClient)

	customLogGroupsToUnMonitor, err := getCustomLogGroups(event.NewIsSecret, event.NewCustom)
	if err != nil {
		sugLog.Error("Error while getting custom log groups: ", err.Error())
		return "", err
	}

	logGroupsToUnMonitor = append(logGroupsToUnMonitor, customLogGroupsToUnMonitor...)
	deleted, err := cwClient.removeSubscriptionFilter(logGroupsToUnMonitor)
	if err != nil {
		sugLog.Error("Error while removing subscription filters: ", err.Error())
		return "", err
	}

	sugLog.Info("Deleted subscription filters for the following log groups: ", deleted)
	return "Event handled successfully", nil
}
