package handler

import (
	"github.com/hashicorp/go-multierror"
	"strings"
	"sync"
)

// getServices returns a list of services to monitor
func getServices() []string {
	servicesStr := envConfig.servicesValue
	if servicesStr == emptyString {
		return nil
	}
	return convertStrToArr(servicesStr)
}

// getCustomGroupsPrefixes returns list of custom log groups which were defined with a wildcard, meaning as prefixes
func getCustomGroupsPrefixes() []string {
	customGroupsStr := envConfig.customGroupsValue
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

// getServicesLogGroups returns a list of log groups to monitor based on the services
func getServicesLogGroups(services []string, cwLogsClient *CloudWatchLogsClient) []string {
	servicesLogGroups := make([]string, 0)
	serviceToPrefix := getServicesMap()
	for _, service := range services {
		if prefix, ok := serviceToPrefix[service]; ok {
			currServiceLG, err := cwLogsClient.getLogGroupsWithPrefix(prefix)
			if err != nil {
				sugLog.Error("Failed to get log groups with prefix: ", prefix)
			}
			servicesLogGroups = append(servicesLogGroups, currServiceLG...)
		}
	}
	return servicesLogGroups
}

// getCustomLogGroups returns a list of custom log groups to monitor
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

// getCustomLogGroupsFromSecret helper function of getCustomLogGroups, returns a list of custom log groups to monitor from secret value
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

// getCustomLogGroupsFromParam helper function of getCustomLogGroups, returns a list of custom log groups to monitor from parameter
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
				newLogGroups, err := cwLogsClient.getLogGroupsWithPrefix(strings.TrimSuffix(logGroup, "*"))
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
