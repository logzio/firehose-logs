package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/logzio/firehose-logs/common"
	"github.com/logzio/firehose-logs/logger"
	"go.uber.org/zap"
	"os"
)

var sugLog *zap.SugaredLogger

func HandleRequest(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	sugLog = logger.GetSugaredLogger()

	sugLog.Info("Starting handling event...")
	sugLog.Debug("Handling event: ", event)

	switch event.RequestType {
	case "Create":
		sugLog.Debug("Detected CloudFormation Create event")
		return createCustomResource(ctx, event)
	case "Update":
		sugLog.Debug("Detected CloudFormation Update event")
		return updateCustomResource(ctx, event)
	case "Delete":
		sugLog.Debug("Detected CloudFormation Delete event")
		return deleteCustomResource(ctx, event)
	default:
		sugLog.Debug("Detected unsupported request type")
		return
	}
}

func createCustomResource(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	physicalResourceID = generatePhysicalResourceId(event)

	payload := common.NewSubscriptionFilterEvent(common.RequestParameters{
		Action:      common.AddSF,
		NewServices: os.Getenv(common.EnvServices),
		NewCustom:   os.Getenv(common.EnvCustomGroups),
		NewIsSecret: os.Getenv(common.EnvSecretEnabled),
	})
	sugLog.Debug("Created SubscriptionFilter Event: ", payload)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		sugLog.Error("Error marshalling payload: ", err.Error())
		return physicalResourceID, nil, err
	}

	stackName := event.ResourceProperties["StackName"].(string)

	err = invokeLambdaAsynchronously(ctx, jsonPayload, stackName)
	if err != nil {
		sugLog.Error("Error invoking lambda: ", err.Error())
		return physicalResourceID, nil, err
	}

	// no need to send back anything to the cfn stack, therefore we return empty map
	return physicalResourceID, make(map[string]interface{}), nil
}

func updateCustomResource(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	physicalResourceID = generatePhysicalResourceId(event)

	oldConfig := event.OldResourceProperties
	newConfig := event.ResourceProperties

	getConfigItem := func(config map[string]interface{}, key string) string {
		if value, ok := config[key].(string); ok {
			return value
		}
		return ""
	}

	payload := common.NewSubscriptionFilterEvent(common.RequestParameters{
		Action:      common.UpdateSF,
		NewServices: getConfigItem(newConfig, common.EnvServices),
		OldServices: getConfigItem(oldConfig, common.EnvServices),
		NewCustom:   getConfigItem(newConfig, common.EnvCustomGroups),
		OldCustom:   getConfigItem(oldConfig, common.EnvCustomGroups),
		NewIsSecret: getConfigItem(newConfig, common.EnvSecretEnabled),
		OldIsSecret: getConfigItem(oldConfig, common.EnvSecretEnabled),
	})
	sugLog.Debug("Created SubscriptionFilter Event: ", payload)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		sugLog.Error("Error marshalling payload: ", err.Error())
		return physicalResourceID, nil, err
	}

	stackName := event.ResourceProperties["StackName"].(string)

	err = invokeLambdaAsynchronously(ctx, jsonPayload, stackName)
	if err != nil {
		sugLog.Error("Error invoking lambda: ", err.Error())
		return physicalResourceID, nil, err
	}

	// no need to send back anything to the cfn stack, therefore we return empty map
	return physicalResourceID, make(map[string]interface{}), nil
}

func deleteCustomResource(ctx context.Context, event cfn.Event) (physicalResourceID string, data map[string]interface{}, err error) {
	physicalResourceID = generatePhysicalResourceId(event)

	payload := common.NewSubscriptionFilterEvent(common.RequestParameters{
		Action:      common.DeleteSF,
		NewServices: os.Getenv(common.EnvServices),
		NewCustom:   os.Getenv(common.EnvCustomGroups),
		NewIsSecret: os.Getenv(common.EnvSecretEnabled),
	})
	sugLog.Debug("Created SubscriptionFilter Event: ", payload)

	jsonPayload, err := json.Marshal(payload)
	if err != nil {
		sugLog.Error("Error marshalling payload: ", err.Error())
		return physicalResourceID, nil, err
	}

	stackName := event.ResourceProperties["StackName"].(string)

	_, err = invokeLambdaSynchronously(ctx, jsonPayload, stackName)
	if err != nil {
		sugLog.Error("Error invoking lambda or executing function: ", err.Error())
		return physicalResourceID, nil, err
	}

	// no need to send back anything to the cfn stack, therefore we return empty map
	return physicalResourceID, make(map[string]interface{}), nil
}

func generatePhysicalResourceId(event cfn.Event) string {
	// Concatenate StackId and LogicalResourceId to form a unique PhysicalResourceId
	physicalResourceId := fmt.Sprintf("%s-%s", event.StackID, event.LogicalResourceID)
	sugLog.Debug("Generated physicalId with value: ", physicalResourceId)
	return physicalResourceId
}
