package handler

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/logzio/firehose-logs/common"
	lp "github.com/logzio/firehose-logs/logger"
	"go.uber.org/zap"
	"os"
	"sort"
	"strings"
)

var sugLog *zap.SugaredLogger

func HandleEventBridgeRequest(ctx context.Context, event map[string]interface{}) (string, error) {
	logger := lp.GetLogger()
	defer logger.Sync()
	sugLog = logger.Sugar()
	sugLog.Info("Starting handling EventBridge event...")
	sugLog.Debug("Handling event: ", event)
	err := common.ValidateRequired()
	if err != nil {
		return "Lambda finished with error", err
	}

	// Extracted EventBridge event handling logic
	if detail, ok := event["detail"].(map[string]interface{}); ok {
		eventName := detail["eventName"]

		switch eventName {
		case "CreateLogGroup":
			if requestParameters, ok := detail["requestParameters"].(map[string]interface{}); ok {
				if logGroupName, ok := requestParameters["logGroupName"].(string); ok {
					newLogGroupCreated(logGroupName)
				} else {
					sugLog.Debug("log group name is not of type string or missing from EventBridge event")
				}
			} else {
				sugLog.Debug("request parameters is not of type map[string]interface{} or missing from EventBridge event.")
			}
		case "PutSecretValue":
			if requestParameters, ok := detail["requestParameters"].(map[string]interface{}); ok {
				if secretId, ok := requestParameters["secretId"].(string); ok {
					secretName := os.Getenv(common.EnvCustomGroups)
					// make sure the secret that changed is the relevant secret
					if secretId == secretName {
						err := updateSecretCustomLogGroups(ctx, secretId)
						if err != nil {
							return "", err
						}
					}
					if strings.Contains(secretId, secretName) {
						err := updateSecretCustomLogGroups(ctx, secretId)
						if err != nil {
							return "", err
						}
					} else {
						sugLog.Debug("The EventBridge event secretId is not the secret that has custom log groups in it. Skipping it.")
					}
				} else {
					sugLog.Debug("secretId is not of string or missing from EventBridge event.")
				}
			} else {
				sugLog.Debug("requestParameters is not of type map[string]interface{} or missing from EventBridge event.")
			}
		default:
			sugLog.Debugf("Detected unsupported event type %s", eventName)
		}
	} else {
		sugLog.Debug("detail is not of type map[string]interface{} or missing from EventBridge event.")
	}

	return "EventBridge event processed", nil
}

func newLogGroupCreated(logGroup string) {
	// Prevent a situation where we put subscription filter on the trigger function
	if logGroup == common.LambdaPrefix+os.Getenv(common.EnvFunctionName) {
		return
	}

	servicesToAdd := common.GetServices()
	var added []string
	if servicesToAdd != nil {
		serviceToPrefix := common.GetServicesMap()
		sess, err := common.GetSession()
		if err != nil {
			sugLog.Error("Could not create aws session: ", err.Error())
			return
		}
		logsClient := cloudwatchlogs.New(sess)
		for _, service := range servicesToAdd {
			if prefix, ok := serviceToPrefix[service]; ok {
				if strings.Contains(logGroup, prefix) {
					added = common.PutSubscriptionFilter([]string{logGroup}, logsClient)
					if len(added) > 0 {
						sugLog.Info("Added log group: ", logGroup)
						return
					}
				}
			}
		}
	}

	sugLog.Info("Log group ", logGroup, " does not match any of the selected services: ", servicesToAdd)
}

func _getLatestOldSecretVersion(ctx context.Context, svc *secretsmanager.Client, secretId string) (*string, error) {
	var latestOldVersionId *string

	listSecretVersionsInput := &secretsmanager.ListSecretVersionIdsInput{
		SecretId: &secretId,
	}

	secretInfo, err := svc.ListSecretVersionIds(ctx, listSecretVersionsInput)
	if err != nil {
		sugLog.Error("Failed to list secret versions to update the monitored custom log groups")
		return nil, err
	}

	// Sort the versions based on created date
	sort.Slice(secretInfo.Versions, func(i, j int) bool {
		return secretInfo.Versions[i].CreatedDate.After(*secretInfo.Versions[j].CreatedDate)
	})

	if len(secretInfo.Versions) > 1 {
		latestOldVersionId = secretInfo.Versions[1].VersionId
	} else {
		sugLog.Warn("Custom log groups secret doesn't have older version to apply changes in comparison to.")
	}

	return latestOldVersionId, nil
}

func _getOldSecretValue(ctx context.Context, svc *secretsmanager.Client, secretId string, oldVersionId *string) (string, error) {
	// get the old version value
	getOldSecretValueInput := &secretsmanager.GetSecretValueInput{
		SecretId:  &secretId,
		VersionId: oldVersionId,
	}

	oldSecret, err := svc.GetSecretValue(ctx, getOldSecretValueInput)
	if err != nil {
		sugLog.Error("Failed to get the old value of the custom log groups secret")
		return "", err
	}
	oldSecretValue := oldSecret.SecretString
	return *oldSecretValue, nil
}

func updateSecretCustomLogGroups(ctx context.Context, secretId string) error {
	// handle the event;  get last version >> update according to it.
	awsConf, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv(common.EnvAwsRegion)))
	if err != nil {
		sugLog.Error("Failed to setup connection to get older custom log groups secret values.")
		return err
	}
	svc := secretsmanager.NewFromConfig(awsConf)

	oldVersionId, err := _getLatestOldSecretVersion(ctx, svc, secretId)
	if err != nil {
		sugLog.Error("Failed to get the older custom log group secret version.")
		return err
	}

	oldSecretValueStr, err := _getOldSecretValue(ctx, svc, secretId, oldVersionId)
	if err != nil {
		sugLog.Error("Failed to get the old custom log group secret version's value.")
		return err
	}

	newSecretValueStr, err := common.GetCustomLogGroups("true", secretId)
	if err != nil {
		sugLog.Error("Failed to get the current custom log group secret value.")
		return err
	}

	oldSecretValue := common.ParseServices(oldSecretValueStr)
	newSecretValue := common.ParseServices(newSecretValueStr)
	customGroupsToAdd, customGroupsToRemove := common.FindDifferences(oldSecretValue, newSecretValue)

	sess, err := common.GetSession()
	if err != nil {
		sugLog.Error("Error while creating session: ", err.Error())
		return err
	}

	if err := common.UpdateSubscriptionFilters(sess, []string{}, []string{}, customGroupsToAdd, customGroupsToRemove); err != nil {
		sugLog.Errorf("Error updating subscription filters: %v", err)
		return err
	}
	return nil
}
