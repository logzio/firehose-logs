package handler

import (
	"context"
	"github.com/aws/aws-lambda-go/cfn"
	lp "github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"testing"
)

func setup(eventType string) (ctx context.Context, event cfn.Event, initLogger *zap.SugaredLogger) {
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
