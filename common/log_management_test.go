package common

import (
	"fmt"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func setup() (logsClient *cloudwatchlogs.CloudWatchLogs) {
	err := os.Setenv(envFirehoseArn, "test-arn")
	if err != nil {
		return
	}

	err = os.Setenv(envAccountId, "aws-account-id")
	if err != nil {
		return
	}

	err = os.Setenv(envAwsPartition, "test-partition")
	if err != nil {
		return
	}

	err = os.Setenv(envAwsRegion, "not-existing-region")
	if err != nil {
		return
	}

	err = os.Setenv(envPutSubscriptionFilterRole, "not-existing-role-arn")
	if err != nil {
		return
	}

	mockSession, _ := GetSession()
	mockClient := cloudwatchlogs.New(mockSession)

	return mockClient
}

func TestValidateRequired(t *testing.T) {
	/* Missing all 3 required env variable */
	result := ValidateRequired()
	assert.Equal(t, "destination ARN must be set", result.Error())

	err := os.Setenv(envFirehoseArn, "test-arn")
	if err != nil {
		return
	}

	/* Missing 2 required env variable */
	result = ValidateRequired()
	assert.Equal(t, "account id must be set", result.Error())

	err = os.Setenv(envAccountId, "aws-account-id")
	if err != nil {
		return
	}

	/* Missing 1 required env variable */
	result = ValidateRequired()
	assert.Equal(t, "aws partition must be set", result.Error())

	/* Valid required env variables */
	err = os.Setenv(envAwsPartition, "test-partition")
	if err != nil {
		return
	}

	result = ValidateRequired()
	assert.Nil(t, result)
}

func TestGetSession(t *testing.T) {
	result, err := GetSession()
	assert.IsType(t, (*session.Session)(nil), result)
	assert.Nil(t, err)
}

func TestPutSubscriptionFilter(t *testing.T) {
	mockClient := setup()
	logGroups := []string{"logGroup1", "logGroup2"}

	added := PutSubscriptionFilter(logGroups, mockClient)
	fmt.Println(added)
}

func TestDeleteSubscriptionFilter(t *testing.T) {}
