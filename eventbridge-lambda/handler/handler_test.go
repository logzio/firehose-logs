package handler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
)

func setup(includeDetail bool) (ctx context.Context, event map[string]interface{}) {
	/* Setup needed env variables */
	err := os.Setenv("FIREHOSE_ARN", "test-arn")
	if err != nil {
		return
	}
	err = os.Setenv("ACCOUNT_ID", "aws-account-id")
	if err != nil {
		return
	}
	err = os.Setenv("AWS_PARTITION", "test-partition")
	if err != nil {
		return
	}

	/* Setup mock context and event */
	mockEvent := map[string]interface{}{
		"version":     "0",
		"id":          "12345678-1234-5678-1234-567812345678",
		"detail-type": "MyCustomEvent",
		"source":      "my.custom.source",
		"account":     "123456789012",
		"time":        "2024-06-20T12:00:00Z",
		"region":      "us-west-2",
		"resources": []string{
			"resource1",
			"resource2",
		},
	}
	if includeDetail {
		mockEvent["detail"] = map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"nestedObject": map[string]interface{}{
				"nestedKey": "nestedValue",
			},
			"requestParameters": map[string]interface{}{
				"logGroupName": "my-log-group",
			},
		}
	}
	ctx = context.Background()

	return ctx, mockEvent
}

func TestHandleNewLogGroupCreatedNoServices(t *testing.T) {
	ctx, mockEvent := setup(true)
	result, err := HandleEventBridgeRequest(ctx, mockEvent)
	assert.Equal(t, "EventBridge event processed", result)
	assert.Nil(t, err)
}

func TestHandleNewLogGroupCreated(t *testing.T) {
	err := os.Setenv("SERVICES", "rds, cloudtrail")
	if err != nil {
		return
	}
	ctx, mockEvent := setup(true)
	result, err := HandleEventBridgeRequest(ctx, mockEvent)
	assert.Equal(t, "EventBridge event processed", result)
	assert.Nil(t, err)
}

func TestHandleNoDetail(t *testing.T) {
	ctx, mockEvent := setup(false)
	result, err := HandleEventBridgeRequest(ctx, mockEvent)
	assert.Equal(t, "EventBridge event processed", result)
	assert.Nil(t, err)
}
