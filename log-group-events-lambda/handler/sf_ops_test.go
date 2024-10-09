package handler

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs/cloudwatchlogsiface"
	lp "github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"sort"
	"testing"
)

type MockCloudWatchLogsClient struct {
	mock.Mock
	cloudwatchlogsiface.CloudWatchLogsAPI
}

func (m *MockCloudWatchLogsClient) PutSubscriptionFilter(input *cloudwatchlogs.PutSubscriptionFilterInput) (*cloudwatchlogs.PutSubscriptionFilterOutput, error) {
	if *input.LogGroupName == "errorGroup" {
		return nil, fmt.Errorf("an error occurred")
	}

	args := m.Called(input)
	return args.Get(0).(*cloudwatchlogs.PutSubscriptionFilterOutput), args.Error(1)
}

func (m *MockCloudWatchLogsClient) DeleteSubscriptionFilter(input *cloudwatchlogs.DeleteSubscriptionFilterInput) (*cloudwatchlogs.DeleteSubscriptionFilterOutput, error) {
	if *input.LogGroupName == "errorGroup" {
		return nil, fmt.Errorf("an error occurred")
	}

	args := m.Called(input)
	return args.Get(0).(*cloudwatchlogs.DeleteSubscriptionFilterOutput), args.Error(1)
}

func (m *MockCloudWatchLogsClient) DescribeLogGroups(input *cloudwatchlogs.DescribeLogGroupsInput) (*cloudwatchlogs.DescribeLogGroupsOutput, error) {
	switch *input.LogGroupNamePrefix {
	case "/aws/apigateway/":
		return &cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{
				{
					LogGroupName: aws.String("/aws/apigateway/g1"),
				}},
		}, nil
	case "/aws/lambda/":
		return &cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{{
				LogGroupName: aws.String("/aws/lambda/g1"),
			},
				{
					LogGroupName: aws.String("/aws/lambda/g2"),
				}},
		}, nil
	case "/aws/codebuild/":
		return &cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{},
		}, nil
	case "/log/group1/":
		return &cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{{
				LogGroupName: aws.String("/log/group1/a"),
			},
				{
					LogGroupName: aws.String("/log/group1/b"),
				}},
		}, nil
	case "/log/group2/":
		return &cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{
				{
					LogGroupName: aws.String("/log/group2/a"),
				}},
		}, nil
	case "/aws/error/test/":
		return &cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{},
		}, fmt.Errorf("an error occurred")
	default:
		return &cloudwatchlogs.DescribeLogGroupsOutput{
			LogGroups: []*cloudwatchlogs.LogGroup{{
				LogGroupName: aws.String("/random/log/group"),
			}},
		}, nil
	}
}

func setupSFTest() {
	/* Setup logger */
	sugLog = lp.GetSugaredLogger()
}

func TestAddSubscriptionFilter(t *testing.T) {
	setupSFTest()

	tests := []struct {
		name          string
		logGroups     []string
		expectedAdded []string
		errorExpected bool
	}{
		{
			name:          "All successful",
			logGroups:     []string{"group1", "group2"},
			expectedAdded: []string{"group1", "group2"},
			errorExpected: false,
		},
		{
			name:          "Error on one group",
			logGroups:     []string{"group1", "errorGroup"},
			expectedAdded: []string{"group1"},
			errorExpected: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockClient := new(MockCloudWatchLogsClient)
			mockClient.On("PutSubscriptionFilter", mock.Anything).Return(&cloudwatchlogs.PutSubscriptionFilterOutput{}, nil).Times(len(test.logGroups))
			mockClient.On("PutSubscriptionFilter", mock.MatchedBy(func(input *cloudwatchlogs.PutSubscriptionFilterInput) bool {
				return *input.LogGroupName == "errorGroup"
			})).Return(nil, fmt.Errorf("an error occurred"))

			cwClient := &CloudWatchLogsClient{Client: mockClient}
			added, err := cwClient.addSubscriptionFilter(test.logGroups)
			sort.Strings(added)

			assert.Equal(t, test.expectedAdded, added, "Expected log groups to be added %v but got %v", test.expectedAdded, added)

			if test.errorExpected {
				assert.NotNil(t, err, "Expected an error but got nil")
			} else {
				assert.Nil(t, err, "Expected error to be nil but got %v", err)
			}
		})
	}
}

func TestUpdateSubscriptionFilters(t *testing.T) {
	setupSFTest()

	tests := []struct {
		name                 string
		servicesToAdd        []string
		servicesToRemove     []string
		customGroupsToAdd    []string
		customGroupsToRemove []string
		errorExpected        bool
	}{
		{
			name:                 "new to add and delete",
			servicesToAdd:        []string{"apigateway", "lambda"},
			servicesToRemove:     []string{"rds"},
			customGroupsToAdd:    []string{"group1"},
			customGroupsToRemove: []string{"group2"},
			errorExpected:        false,
		},
		{
			name:                 "no new to add",
			servicesToAdd:        []string{},
			servicesToRemove:     []string{"lambda"},
			customGroupsToAdd:    []string{},
			customGroupsToRemove: []string{"group2"},
			errorExpected:        false,
		},
		{
			name:                 "no new to delete",
			servicesToAdd:        []string{"apigateway", "rds"},
			servicesToRemove:     []string{},
			customGroupsToAdd:    []string{"group1"},
			customGroupsToRemove: []string{},
			errorExpected:        false,
		},
		{
			name:                 "error in add",
			servicesToAdd:        []string{},
			servicesToRemove:     []string{},
			customGroupsToAdd:    []string{"errorGroup"},
			customGroupsToRemove: []string{},
			errorExpected:        true,
		},
		{
			name:                 "error in delete",
			servicesToAdd:        []string{},
			servicesToRemove:     []string{},
			customGroupsToAdd:    []string{},
			customGroupsToRemove: []string{"errorGroup"},
			errorExpected:        true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockClient := new(MockCloudWatchLogsClient)
			mockClient.On("PutSubscriptionFilter", mock.Anything).Return(&cloudwatchlogs.PutSubscriptionFilterOutput{}, nil)
			mockClient.On("PutSubscriptionFilter", mock.MatchedBy(func(input *cloudwatchlogs.PutSubscriptionFilterInput) bool {
				return *input.LogGroupName == "errorGroup"
			})).Return(nil, fmt.Errorf("an error occurred"))

			mockClient.On("DeleteSubscriptionFilter", mock.Anything).Return(&cloudwatchlogs.DeleteSubscriptionFilterOutput{}, nil)
			mockClient.On("DeleteSubscriptionFilter", mock.MatchedBy(func(input *cloudwatchlogs.DeleteSubscriptionFilterInput) bool {
				return *input.LogGroupName == "errorGroup"
			})).Return(nil, fmt.Errorf("an error occurred"))

			cwClient := &CloudWatchLogsClient{Client: mockClient}
			err := cwClient.updateSubscriptionFilters(test.servicesToAdd, test.servicesToRemove, test.customGroupsToAdd, test.customGroupsToRemove)

			if test.errorExpected {
				assert.NotNil(t, err, "Expected an error but got nil")
			} else {
				assert.Nil(t, err, "Expected error to be nil but got %v", err)
			}
		})
	}
}

func TestDeleteSubscriptionFilters(t *testing.T) {
	setupSFTest()

	tests := []struct {
		name           string
		logGroups      []string
		expectedRemove []string
		errorExpected  bool
	}{
		{
			name:           "All successful",
			logGroups:      []string{"group1", "group2"},
			expectedRemove: []string{"group1", "group2"},
			errorExpected:  false,
		},
		{
			name:           "error on one group",
			logGroups:      []string{"group1", "errorGroup"},
			expectedRemove: []string{"group1"},
			errorExpected:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockClient := new(MockCloudWatchLogsClient)
			mockClient.On("DeleteSubscriptionFilter", mock.Anything).Return(&cloudwatchlogs.DeleteSubscriptionFilterOutput{}, nil).Times(len(test.logGroups))
			mockClient.On("DeleteSubscriptionFilter", mock.MatchedBy(func(input *cloudwatchlogs.DeleteSubscriptionFilterInput) bool {
				return *input.LogGroupName == "errorGroup"
			})).Return(nil, fmt.Errorf("an error occurred"))

			cwClient := &CloudWatchLogsClient{Client: mockClient}
			removed, err := cwClient.removeSubscriptionFilter(test.logGroups)
			sort.Strings(removed)

			assert.Equal(t, test.expectedRemove, removed, "Expected log groups to be removed %v but got %v", test.expectedRemove, removed)

			if test.errorExpected {
				assert.NotNil(t, err, "Expected an error but got nil")
			} else {
				assert.Nil(t, err, "Expected error to be nil but got %v", err)
			}
		})
	}
}
