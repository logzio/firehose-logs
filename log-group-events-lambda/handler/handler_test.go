package handler

import (
	"context"
	"os"
	"testing"

	"github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
)

func setupHandlerTest() (ctx context.Context) {
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

	ctx = context.Background()

	return ctx
}

func TestUnsupportedEventHandling(t *testing.T) {
	ctx := setupHandlerTest()

	tests := []struct {
		name              string
		event             map[string]interface{}
		expectedOutputMsg string
		expectedError     bool
	}{
		{
			name: "Unsupported event with all required fields",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "MyCustomEvent",
					"requestParameters": map[string]interface{}{
						"logGroupName": "my-log-group",
					},
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name:              "Unsupported event with missing detail field",
			event:             map[string]interface{}{},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name: "Unsupported event with missing eventName field",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"requestParameters": map[string]interface{}{
						"logGroupName": "my-log-group",
					},
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name: "Unsupported event with missing requestParameters field",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "MyCustomEvent",
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name: "CreateLogGroup event with missing logGroup field",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName":         "CreateLogGroup",
					"requestParameters": map[string]interface{}{},
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name: "PutSecretValue event with missing secretId field",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName":         "PutSecretValue",
					"requestParameters": map[string]interface{}{},
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := HandleRequest(ctx, test.event)

			assert.NotNil(t, err)
			assert.Equal(t, test.expectedOutputMsg, res)
		})
	}
}

func TestTagResourceEventHandling(t *testing.T) {
	ctx := setupHandlerTest()

	tests := []struct {
		name              string
		event             map[string]interface{}
		expectedOutputMsg string
		expectedError     bool
	}{
		{
			name: "TagResource event for CloudWatch Log Group with valid ARN",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "TagResource",
					"requestParameters": map[string]interface{}{
						"resourceArn": "arn:aws:logs:us-east-1:123456789012:log-group:/aws/lambda/my-function",
						"tags": map[string]interface{}{
							"LogzIO": "True",
						},
					},
				},
			},
			expectedOutputMsg: "TagResource event handled successfully",
			expectedError:     false,
		},
		{
			name: "TagResource event with missing resourceArn field",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "TagResource",
					"requestParameters": map[string]interface{}{
						"tags": map[string]interface{}{
							"LogzIO": "True",
						},
					},
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name: "TagResource event with invalid ARN format",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "TagResource",
					"requestParameters": map[string]interface{}{
						"resourceArn": "not-an-arn",
						"tags": map[string]interface{}{
							"LogzIO": "True",
						},
					},
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name: "TagResource20170331v2 event for Lambda with valid ARN",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "TagResource20170331v2",
					"requestParameters": map[string]interface{}{
						"resource": "arn:aws:lambda:us-east-1:123456789012:function:my-payment-processor",
						"tags": map[string]interface{}{
							"LogzIO": "True",
						},
					},
				},
			},
			expectedOutputMsg: "TagResource20170331v2 event handled successfully",
			expectedError:     false,
		},
		{
			name: "TagResource20170331v2 event with missing resource field",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "TagResource20170331v2",
					"requestParameters": map[string]interface{}{
						"tags": map[string]interface{}{
							"LogzIO": "True",
						},
					},
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name: "TagResource20170331v2 event with invalid Lambda ARN",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "TagResource20170331v2",
					"requestParameters": map[string]interface{}{
						"resource": "invalid-lambda-arn",
						"tags": map[string]interface{}{
							"LogzIO": "True",
						},
					},
				},
			},
			expectedOutputMsg: "",
			expectedError:     true,
		},
		{
			name: "TagResource event for Log Group with colon in name",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "TagResource",
					"requestParameters": map[string]interface{}{
						"resourceArn": "arn:aws:logs:us-east-1:123456789012:log-group:/aws/lambda/my-function:with:colons",
						"tags": map[string]interface{}{
							"LogzIO": "True",
						},
					},
				},
			},
			expectedOutputMsg: "TagResource event handled successfully",
			expectedError:     false,
		},
		{
			name: "TagResource20170331v2 event for Lambda with colon in function name",
			event: map[string]interface{}{
				"detail": map[string]interface{}{
					"eventName": "TagResource20170331v2",
					"requestParameters": map[string]interface{}{
						"resource": "arn:aws:lambda:us-east-1:123456789012:function:my-function:alias",
						"tags": map[string]interface{}{
							"LogzIO": "True",
						},
					},
				},
			},
			expectedOutputMsg: "TagResource20170331v2 event handled successfully",
			expectedError:     false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := HandleRequest(ctx, test.event)

			if test.expectedError {
				assert.NotNil(t, err)
				assert.Equal(t, test.expectedOutputMsg, res)
			} else {
				assert.Nil(t, err)
				assert.Equal(t, test.expectedOutputMsg, res)
			}
		})
	}
}

func TestHandleTagResourceEvent(t *testing.T) {
	ctx := setupHandlerTest()
	
	// Initialize logger and config (normally done in HandleRequest)
	sugLog = logger.GetSugaredLogger()
	envConfig = NewConfig()

	tests := []struct {
		name             string
		taggedResource   string
		expectedError    bool
		expectedErrorMsg string
	}{
		{
			name:           "Valid CloudWatch Log Group ARN",
			taggedResource: "arn:aws:logs:us-east-1:123456789012:log-group:/aws/lambda/my-function",
			expectedError:  false,
		},
		{
			name:           "Valid Lambda Function ARN",
			taggedResource: "arn:aws:lambda:us-east-1:123456789012:function:my-payment-processor",
			expectedError:  false,
		},
		{
			name:             "Invalid ARN format - not an ARN",
			taggedResource:   "not-an-arn",
			expectedError:    true,
			expectedErrorMsg: "provided string is not AWS arn",
		},
		{
			name:             "Invalid ARN format - missing parts",
			taggedResource:   "arn:aws:logs",
			expectedError:    true,
			expectedErrorMsg: "provided string is not AWS arn",
		},
		{
			name:           "Valid Log Group ARN with colons in name",
			taggedResource: "arn:aws:logs:us-east-1:123456789012:log-group:/custom/path/with:colons:here",
			expectedError:  false,
		},
		{
			name:           "Valid Lambda ARN with alias",
			taggedResource: "arn:aws:lambda:us-east-1:123456789012:function:my-function:prod",
			expectedError:  false,
		},
		{
			name:             "Invalid resource type - S3",
			taggedResource:   "arn:aws:s3:::my-bucket/my-key",
			expectedError:    true,
			expectedErrorMsg: "unable to get name from arn.resource ",
		},
		{
			name:             "Invalid resource type - EC2",
			taggedResource:   "arn:aws:ec2:us-east-1:123456789012:instance/i-1234567890abcdef0",
			expectedError:    true,
			expectedErrorMsg: "unable to get name from arn.resource ",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			res, err := handleTagResourceEvent(ctx, test.taggedResource)

			if test.expectedError {
				assert.NotNil(t, err)
				if test.expectedErrorMsg != "" {
					assert.Contains(t, err.Error(), test.expectedErrorMsg)
				}
			} else {
				assert.Nil(t, err)
				assert.Equal(t, "Tag Resource Event handled successfully.", res)
			}
		})
	}
}
