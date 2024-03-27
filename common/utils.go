package common

import (
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

func GetCustomPaths() []string {
	pathsStr := os.Getenv(envCustomGroups)
	if pathsStr == emptyString {
		return nil
	}

	pathsStr = strings.ReplaceAll(pathsStr, " ", "")
	return strings.Split(pathsStr, valuesSeparator)
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
