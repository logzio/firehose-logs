package handler

import (
	"fmt"
	"os"
	"testing"

	"github.com/logzio/firehose-logs/common"
	"github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
)

func InitConfigTest() {
	sugLog = logger.GetSugaredLogger()
}

func TestNewConfigMissingEnv(t *testing.T) {
	InitConfigTest()
	conf := NewConfig()
	assert.Nil(t, conf)
}

func TestNewConfigValidRequired(t *testing.T) {
	InitConfigTest()

	/* Setup required env variable */
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

	conf := NewConfig()
	assert.NotNil(t, conf)
	assert.Equal(t, "test-arn", conf.destinationArn)
	assert.Equal(t, "aws-account-id", conf.accountId)
	assert.Equal(t, "", conf.region)
	assert.Equal(t, "/aws/lambda/", conf.thisFunctionLogGroup)
	assert.Equal(t, "", conf.thisFunctionName)
	assert.Equal(t, "", conf.customGroupsValue)
	assert.Equal(t, "", conf.servicesValue)
}

func TestNewConfigValid(t *testing.T) {
	InitConfigTest()

	/* Setup env variable */
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

	err = os.Setenv(envFunctionName, "g2")
	if err != nil {
		return
	}

	fmt.Println("test3")

	conf := NewConfig()
	assert.NotNil(t, conf)
	assert.Equal(t, "test-arn", conf.destinationArn)
	assert.Equal(t, "aws-account-id", conf.accountId)
	assert.Equal(t, "/aws/lambda/g2", conf.thisFunctionLogGroup)
	assert.Equal(t, "g2", conf.thisFunctionName)
	assert.Equal(t, "", conf.customGroupsValue)
	assert.Equal(t, "", conf.servicesValue)
	assert.Equal(t, "", conf.region)
}

func TestValidateRequired(t *testing.T) {
	/* Setup tests */
	InitConfigTest()

	tests := []struct {
		name          string
		conf          Config
		expectedError bool
		errorStr      string
	}{
		{
			name: "missing all 3 required",
			conf: Config{
				awsPartition:   "",
				destinationArn: "",
				accountId:      "",
			},
			expectedError: true,
			errorStr:      "destination ARN must be set",
		},
		{
			name: "missing 2 required",
			conf: Config{
				awsPartition:   "partition",
				destinationArn: "",
				accountId:      "",
			},
			expectedError: true,
			errorStr:      "destination ARN must be set",
		},
		{
			name: "missing 1 required",
			conf: Config{
				awsPartition:   "partition",
				destinationArn: "some-arn",
				accountId:      "",
			},
			expectedError: true,
			errorStr:      "account id must be set",
		},
		{
			name: "valid",
			conf: Config{
				awsPartition:   "partition",
				destinationArn: "some-arn",
				accountId:      "accountId",
			},
			expectedError: false,
			errorStr:      "",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := test.conf.validateRequired()
			if test.expectedError {
				assert.NotNil(t, result)
				assert.Equal(t, test.errorStr, result.Error())
			} else {
				assert.Nil(t, result)
			}
		})

	}
}

func TestValidateFilterPattern(t *testing.T) {
	InitConfigTest()

	err := os.Setenv(common.EnvAwsRegion, "us-east-1")
	if err != nil {
		t.Fatalf("Failed to set AWS region: %v", err)
	}

	tests := []struct {
		name          string
		filterPattern string
		expectedError bool
		errorContains string
	}{
		{
			name:          "empty pattern",
			filterPattern: "",
			expectedError: false,
			errorContains: "",
		},
		{
			name:          "valid simple pattern",
			filterPattern: "[ip, user_id, username]",
			expectedError: false,
			errorContains: "",
		},
		{
			name:          "valid pattern with wildcard",
			filterPattern: "[..., status_code=404, size]",
			expectedError: false,
			errorContains: "",
		},
		{
			name:          "valid text pattern",
			filterPattern: "\"ERROR\"",
			expectedError: false,
			errorContains: "",
		},
		{
			name:          "valid complex pattern",
			filterPattern: "[timestamp, request_id, level=ERROR, message]",
			expectedError: false,
			errorContains: "",
		},
		{
			name:          "unbalanced brackets",
			filterPattern: "[ip, user_id",
			expectedError: true,
			errorContains: "invalid filter pattern",
		},
		{
			name:          "invalid syntax",
			filterPattern: "[timestamp=, request_id, level=ERROR]",
			expectedError: true,
			errorContains: "invalid filter pattern",
		},
		{
			name:          "unbalanced quotes",
			filterPattern: "\"ERROR",
			expectedError: true,
			errorContains: "invalid filter pattern",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			config := &Config{
				region:        "us-east-1",
				filterPattern: test.filterPattern,
			}

			err := config.validateFilterPattern()

			if test.expectedError {
				assert.Error(t, err)
				if test.errorContains != "" {
					assert.Contains(t, err.Error(), test.errorContains)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
