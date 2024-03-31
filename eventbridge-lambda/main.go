package main

import (
	"context"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/logzio/firehose-logs/common"
	lp "github.com/logzio/firehose-logs/logger"
	"go.uber.org/zap"
	"os"
	"strings"
)

var sugLog *zap.SugaredLogger

func main() {
	lambda.Start(HandleEventBridgeRequest)
}

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

	if _, ok := event["detail"]; ok {
		// Extracted EventBridge event handling logic
		newLogGroupCreated(event["detail"].(map[string]interface{})["requestParameters"].(map[string]interface{})["logGroupName"].(string))
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
