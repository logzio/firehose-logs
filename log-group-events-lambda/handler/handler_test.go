package handler

import (
	"context"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
