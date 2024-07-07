package common

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"os"
	"strings"
)

const (
	envServices                  = "SERVICES"
	envAwsRegion                 = "AWS_REGION"               // reserved env
	EnvFunctionName              = "AWS_LAMBDA_FUNCTION_NAME" // reserved env
	envFirehoseArn               = "FIREHOSE_ARN"
	envAccountId                 = "ACCOUNT_ID"
	envCustomGroups              = "CUSTOM_GROUPS"
	envSecretEnabled             = "SECRET_ENABLED"
	envAwsPartition              = "AWS_PARTITION"
	envPutSubscriptionFilterRole = "PUT_SF_ROLE"

	valuesSeparator        = ","
	emptyString            = ""
	LambdaPrefix           = "/aws/lambda/"
	subscriptionFilterName = "logzio_firehose"
)

func GetServices() []string {
	servicesStr := os.Getenv(envServices)
	if servicesStr == emptyString {
		return nil
	}

	servicesStr = strings.ReplaceAll(servicesStr, " ", "")
	return strings.Split(servicesStr, valuesSeparator)
}

func GetServicesMap() map[string]string {
	return map[string]string{
		"apigateway":       "/aws/apigateway/",
		"rds":              "/aws/rds/",
		"cloudhsm":         "/aws/cloudhsm/",
		"cloudtrail":       "aws-cloudtrail-logs-",
		"codebuild":        "/aws/codebuild/",
		"connect":          "/aws/connect/",
		"elasticbeanstalk": "/aws/elasticbeanstalk/",
		"ecs":              "/aws/ecs/",
		"eks":              "/aws/eks/",
		"aws-glue":         "/aws/aws-glue/",
		"aws-iot":          "AWSIotLogsV2",
		"lambda":           "/aws/lambda/",
		"macie":            "/aws/macie/",
		"amazon-mq":        "/aws/amazonmq/broker/",
	}
}

func GetCustomLogGroups(secretEnabled, customLogGroupsPrmVal string) (string, error) {
	initLogger()
	if secretEnabled == "true" {
		sugLog.Debug("Attempting to get custom log groups from secret parameter: ", customLogGroupsPrmVal)
		secretCache, err := secretcache.New()
		if err != nil {
			return "", err
		}

		result, err := secretCache.GetSecretString(customLogGroupsPrmVal)
		if err != nil {
			return "", err
		}

		var secretValues map[string]string
		err = json.Unmarshal([]byte(result), &secretValues)
		if err != nil {
			return "", err
		}

		customLogGroupsSecret, ok := secretValues["logzioCustomLogGroups"]
		if !ok {
			return "", fmt.Errorf("did not find logzioCustomLogGroups key in the secret %s", customLogGroupsPrmVal)
		}
		return customLogGroupsSecret, nil
	}

	return customLogGroupsPrmVal, nil
}

func GetCustomPaths() []string {
	initLogger()
	pathsStr := os.Getenv(envCustomGroups)
	secretEnabled := os.Getenv(envSecretEnabled)
	if pathsStr == emptyString {
		return nil
	}
	sugLog.Debug("Getting custom log groups with information; secret enabled: ", secretEnabled)
	customLogGroupsStr, err := GetCustomLogGroups(secretEnabled, pathsStr)
	if err != nil {
		sugLog.Errorf("Failed to get custom log groups from secret due to %s", err.Error())
		return nil
	}

	customLogGroupsStr = strings.ReplaceAll(customLogGroupsStr, " ", "")
	return strings.Split(customLogGroupsStr, valuesSeparator)
}

func ParseServices(servicesStr string) []string {
	if servicesStr == emptyString {
		return nil
	}

	servicesStr = strings.ReplaceAll(servicesStr, " ", "")
	return strings.Split(servicesStr, valuesSeparator)
}

func ListContains(s string, l []string) bool {
	for _, item := range l {
		if s == item {
			return true
		}
	}

	return false
}
