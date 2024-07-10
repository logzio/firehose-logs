package common

import (
	"encoding/json"
	"fmt"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"os"
	"regexp"
	"strings"
)

const (
	envServices                  = "SERVICES"
	EnvAwsRegion                 = "AWS_REGION"               // reserved env
	EnvFunctionName              = "AWS_LAMBDA_FUNCTION_NAME" // reserved env
	envFirehoseArn               = "FIREHOSE_ARN"
	envAccountId                 = "ACCOUNT_ID"
	EnvCustomGroups              = "CUSTOM_GROUPS"
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

// FindDifferences finds elements in 'new' that are not in 'old', and vice versa.
func FindDifferences(old, new []string) (toAdd, toRemove []string) {
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

func GetSecretNameFromArn(secretArn string) string {
	var secretName string
	if sugLog == nil {
		initLogger()
	}

	sugLog.Debugf("Attempting to extract secret name from ARN: '%s'", secretArn)

	getSecretName := regexp.MustCompile(fmt.Sprintf(`^arn:aws:secretsmanager:%s:%s:secret:(?P<secretName>\S+)-`, os.Getenv(EnvAwsRegion), os.Getenv(envAccountId)))
	match := getSecretName.FindStringSubmatch(secretArn)

	for i, key := range getSecretName.SubexpNames() {
		if key == "secretName" && len(match) > i {
			secretName = match[i]
			break
		}
	}

	return secretName
}

func GetCustomLogGroups(secretEnabled, customLogGroupsPrmVal string) (string, error) {
	if sugLog == nil {
		initLogger()
	}
	if secretEnabled == "true" {
		sugLog.Debug("Attempting to get custom log groups from secret parameter: ", customLogGroupsPrmVal)
		secretCache, err := secretcache.New()
		if err != nil {
			return "", err
		}

		secretName := GetSecretNameFromArn(customLogGroupsPrmVal)

		result, err := secretCache.GetSecretString(secretName)
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
	if sugLog == nil {
		initLogger()
	}
	pathsStr := os.Getenv(EnvCustomGroups)
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
