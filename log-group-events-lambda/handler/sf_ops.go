package handler

import (
	"errors"
	"fmt"
	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	"github.com/hashicorp/go-multierror"
	"github.com/logzio/firehose-logs/common"
	"os"
	"sync"
	"time"
)

type CloudWatchLogsClient struct {
	Client cloudwatchlogsiface.CloudWatchLogsAPI
}

func getCloudWatchLogsClient() (*CloudWatchLogsClient, error) {
	sess, err := common.GetSession()
	if err != nil {
		return nil, err
	}
	return &CloudWatchLogsClient{Client: cloudwatchlogs.New(sess)}, nil
}

func (cwLogsClient *CloudWatchLogsClient) addSubscriptionFilter(logGroups []string) ([]string, error) {
	if cwLogsClient == nil {
		return nil, fmt.Errorf("CloudWatch Logs client is nil")
	}

	destinationArn := os.Getenv(envFirehoseArn)
	roleArn := os.Getenv(envPutSubscriptionFilterRole)
	filterPattern := ""
	filterName := subscriptionFilterName
	added := make([]string, 0, len(logGroups))
	var result *multierror.Error

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, logGroup := range logGroups {
		// Prevent a situation where we put subscription filter on the trigger function
		if logGroup == lambdaPrefix+os.Getenv(envFunctionName) {
			continue
		}

		wg.Add(1)
		go func(logGroup string) {
			defer wg.Done()

			retries := 0
			for {
				_, err := cwLogsClient.Client.PutSubscriptionFilter(&cloudwatchlogs.PutSubscriptionFilterInput{
					DestinationArn: &destinationArn,
					FilterName:     &filterName,
					LogGroupName:   &logGroup,
					FilterPattern:  &filterPattern,
					RoleArn:        &roleArn,
				})

				// retry mechanism
				if err != nil {
					var awsErr awserr.Error
					ok := errors.As(err, &awsErr)
					if ok && awsErr.Code() == "ThrottlingException" && retries < maxRetries {
						time.Sleep(time.Second * time.Duration(retries*retries))
						retries++
						continue
					} else {
						sugLog.Errorf("Error while trying to add subscription filter for %s: %v", logGroup, err.Error())
						result = multierror.Append(result, err)
						return
					}
				}
				mu.Lock()
				added = append(added, logGroup)
				mu.Unlock()
			}
		}(logGroup)
	}
	wg.Wait()

	return added, result.ErrorOrNil()
}

func (cwLogsClient *CloudWatchLogsClient) updateSubscriptionFilters(servicesToAdd, servicesToRemove, customGroupsToAdd, customGroupsToRemove []string) error {
	var result *multierror.Error

	logGroupsToMonitor := getServicesLogGroups(servicesToAdd, cwLogsClient)
	logGroupsToMonitor = append(logGroupsToMonitor, customGroupsToAdd...)

	if len(logGroupsToMonitor) > 0 {
		added, err := cwLogsClient.addSubscriptionFilter(logGroupsToMonitor)
		if err != nil {
			result = multierror.Append(result, err)
		}
		sugLog.Info("Added subscription filters for the following log groups: ", added)
	} else {
		sugLog.Debug("No new log groups to monitor")
	}

	logGroupsToUnMonitor := getServicesLogGroups(servicesToRemove, cwLogsClient)
	logGroupsToUnMonitor = append(logGroupsToUnMonitor, customGroupsToRemove...)

	if len(logGroupsToUnMonitor) > 0 {
		deleted, err := cwLogsClient.removeSubscriptionFilter(logGroupsToUnMonitor)
		if err != nil {
			result = multierror.Append(result, err)
		}
		sugLog.Info("Deleted subscription filters for the following log groups: ", deleted)
	} else {
		sugLog.Debug("No log groups to stop monitoring")
	}

	return result.ErrorOrNil()
}

func (cwLogsClient *CloudWatchLogsClient) removeSubscriptionFilter(logGroups []string) ([]string, error) {
	if cwLogsClient == nil {
		return nil, fmt.Errorf("CloudWatch Logs client is nil")
	}

	filterName := subscriptionFilterName
	deleted := make([]string, 0, len(logGroups))
	var result *multierror.Error

	var wg sync.WaitGroup
	var mu sync.Mutex

	for _, logGroup := range logGroups {
		wg.Add(1)
		go func(logGroup string) {
			defer wg.Done()

			retries := 0
			for {
				_, err := cwLogsClient.Client.DeleteSubscriptionFilter(&cloudwatchlogs.DeleteSubscriptionFilterInput{
					FilterName:   &filterName,
					LogGroupName: &logGroup,
				})

				// retry mechanism
				if err != nil {
					var awsErr awserr.Error
					ok := errors.As(err, &awsErr)
					if ok && awsErr.Code() == "ThrottlingException" && retries < maxRetries {
						time.Sleep(time.Second * time.Duration(retries*retries))
						retries++
						continue
					} else {
						sugLog.Errorf("Error while trying to delete subscription filter for %s: %v", logGroup, err.Error())
						result = multierror.Append(result, err)
						return
					}
				}
				mu.Lock()
				deleted = append(deleted, logGroup)
				mu.Unlock()
			}
		}(logGroup)
	}
	wg.Wait()

	return deleted, result.ErrorOrNil()
}