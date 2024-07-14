package handler

import (
	"context"
	"github.com/aws/aws-lambda-go/cfn"
	lp "github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"os"
	"testing"
)

func stringPtr(s string) *string {
	/* helper function */
	return &s
}

func setup(eventType string) (ctx context.Context, event cfn.Event, initLogger *zap.SugaredLogger) {
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
	mockEvent := cfn.Event{
		RequestType:        cfn.RequestType(eventType),
		ResponseURL:        "http://pre-signed-S3-url-for-response",
		StackID:            "arn:aws:cloudformation:us-west-2:EXAMPLE/stack-name/guid",
		RequestID:          "unique id for this create request",
		LogicalResourceID:  "MyTestResource",
		PhysicalResourceID: "MyTestResourceId",
		ResourceType:       "Custom::TestResource",
		ResourceProperties: map[string]interface{}{
			"Name": "Test",
		},
		OldResourceProperties: map[string]interface{}{},
	}
	ctx = context.Background()

	/* Setup logger */
	logger := lp.GetLogger()
	defer logger.Sync()
	sugLog = logger.Sugar()

	return ctx, mockEvent, sugLog
}

func TestUnsupportedEventHandling(t *testing.T) {
	ctx, mockEvent, _ := setup("Random")
	res, data, err := HandleRequest(ctx, mockEvent)
	assert.Empty(t, res)
	assert.Empty(t, data)
	assert.Nil(t, err)
}

func TestGeneratePhysicalResourceId(t *testing.T) {
	_, mockEvent, _ := setup("Create")
	physicalId := generatePhysicalResourceId(mockEvent)
	assert.Equal(t, "arn:aws:cloudformation:us-west-2:EXAMPLE/stack-name/guid-MyTestResource", physicalId)
}

func TestCustomResourceRunUpdate(t *testing.T) {
	ctx, mockEvent, sugLog := setup("Create")
	sugLog.Info("init susLog")

	physicalId, data, err := customResourceRunUpdate(ctx, mockEvent)
	assert.Equal(t, "arn:aws:cloudformation:us-west-2:EXAMPLE/stack-name/guid-MyTestResource", physicalId)
	assert.Empty(t, data)
	assert.Nil(t, err)
}

func TestUpdateConfiguration(t *testing.T) {
	tests := []struct {
		name      string
		oldConfig map[string]interface{}
		newConfig map[string]interface{}
		expected  *string
	}{
		{
			name:      "invalid configuration value",
			oldConfig: map[string]interface{}{"Services": 123},
			newConfig: map[string]interface{}{"Services": "elb", "CustomLogGroups": "rand"},
			expected:  stringPtr("invalid configuration type for Services"),
		},
		{
			name:      "no valid log groups configuration",
			oldConfig: map[string]interface{}{},
			newConfig: map[string]interface{}{"Services": "elb", "CustomLogGroups": "rand"},
			expected:  stringPtr("Could not retrieve any log groups"),
		},
		{
			name:      "no changes",
			oldConfig: map[string]interface{}{"Services": "rds", "CustomLogGroups": "rand"},
			newConfig: map[string]interface{}{"Services": "rds", "CustomLogGroups": "rand"},
			expected:  nil,
		},
	}

	ctx, _, _ := setup("Update")

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := updateConfiguration(ctx, test.oldConfig, test.newConfig)

			if test.expected == nil {
				assert.Nil(t, result, "Expected nil, got %v", result)
			} else {
				assert.Equal(t, *test.expected, result.Error(), "Expected %v, got %v", *test.expected, result)
			}
		})
	}
}
