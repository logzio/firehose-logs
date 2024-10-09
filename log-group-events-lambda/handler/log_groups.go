package handler

import (
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/hashicorp/go-multierror"
	"github.com/logzio/firehose-logs/common"
	"os"
	"strings"
	"sync"
)

func getServices() []string {
	servicesStr := os.Getenv(common.EnvServices)
	if servicesStr == emptyString {
		return nil
	}
	return convertStrToArr(servicesStr)
}

func getCustomGroupsPrefixes() []string {
	customGroupsStr := os.Getenv(common.EnvCustomGroups)
	if customGroupsStr == emptyString {
		return nil
	}
	customGroups := convertStrToArr(customGroupsStr)
	var wg sync.WaitGroup
	var mu sync.Mutex

	prefixes := make([]string, 0)
	for _, logGroup := range customGroups {
		wg.Add(1)
		go func(logGroup string) {
			defer wg.Done()

			if strings.HasSuffix(logGroup, "*") {
				mu.Lock()
				prefixes = append(prefixes, strings.TrimSuffix(logGroup, "*"))
				mu.Unlock()
			}
		}(logGroup)
	}
	wg.Wait()

	return prefixes
}

func getServicesLogGroups(services []string, cwLogsClient *CloudWatchLogsClient) []string {
	servicesLogGroups := make([]string, 0)
	serviceToPrefix := getServicesMap()
	for _, service := range services {
		if prefix, ok := serviceToPrefix[service]; ok {
			currServiceLG, err := getLogGroupsWithPrefix(prefix, cwLogsClient)
			if err != nil {
				sugLog.Error("Failed to get log groups with prefix: ", prefix)
			}
			servicesLogGroups = append(servicesLogGroups, currServiceLG...)
		}
	}
	return servicesLogGroups
}

func getLogGroupsWithPrefix(prefix string, cwLogsClient *CloudWatchLogsClient) ([]string, error) {
	var nextToken *string
	logGroups := make([]string, 0)
	for {
		describeOutput, err := cwLogsClient.Client.DescribeLogGroups(&cloudwatchlogs.DescribeLogGroupsInput{
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
				if *logGroup.LogGroupName != lambdaPrefix+os.Getenv(envFunctionName) {
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

func getCustomLogGroups(secretEnabled, customLogGroupsPrmVal string) ([]string, error) {
	cwLogsClient, err := getCloudWatchLogsClient()
	if err != nil {
		sugLog.Error("Failed to get cloudwatch logs client")
	}

	if secretEnabled == "true" {
		secretCache, err := getSecretCacheClient()
		if err != nil {
			sugLog.Error("Failed to get secret cache client")
			return nil, err
		}
		return getCustomLogGroupsFromSecret(customLogGroupsPrmVal, secretCache, cwLogsClient)
	}

	return getCustomLogGroupsFromParam(convertStrToArr(customLogGroupsPrmVal), cwLogsClient)
}

func getCustomLogGroupsFromSecret(secretArn string, secretCache *SecretCacheClient, cwLogsClient *CloudWatchLogsClient) ([]string, error) {
	secretName := getSecretNameFromArn(secretArn)

	secretStruct, err := secretCache.Client.GetSecretString(secretName)
	if err != nil {
		sugLog.Error("Error while getting secret value from cache.")
		return nil, err
	}

	customLogGroups, err := extractCustomGroupsFromSecret(secretArn, secretStruct)
	if err != nil {
		sugLog.Error("Error while extracting custom log groups from secret: ", err.Error())
		return nil, err
	}

	if cwLogsClient != nil {
		sugLog.Warn("Missing CloudWatch logs client, will not handle custom log group names with wildcards.")
		return getCustomLogGroupsFromParam(convertStrToArr(customLogGroups), cwLogsClient)
	}
	return convertStrToArr(customLogGroups), nil
}

func getCustomLogGroupsFromParam(logGroups []string, cwLogsClient *CloudWatchLogsClient) ([]string, error) {
	customLogGroups := make([]string, 0, len(logGroups))
	var result *multierror.Error
	var wg sync.WaitGroup
	var mu sync.Mutex

	if cwLogsClient == nil {
		// we shouldn't fail the entire process only if the cwLogsClient failed to get created
		return logGroups, nil
	}

	for i, logGroup := range logGroups {
		wg.Add(1)
		go func(i int, logGroup string) {
			defer wg.Done()

			if strings.HasSuffix(logGroup, "*") {
				newLogGroups, err := getLogGroupsWithPrefix(strings.TrimSuffix(logGroup, "*"), cwLogsClient)
				if err != nil {
					sugLog.Error("Failed to get log groups with prefix: ", logGroup)
					result = multierror.Append(result, err)
					return
				}
				mu.Lock()
				customLogGroups = append(customLogGroups, newLogGroups...)
				mu.Unlock()

			} else {
				mu.Lock()
				customLogGroups = append(customLogGroups, logGroup)
				mu.Unlock()
			}
		}(i, logGroup)
	}
	wg.Wait()
	return customLogGroups, result.ErrorOrNil()
}
