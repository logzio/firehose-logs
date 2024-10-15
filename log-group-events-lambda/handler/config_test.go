package handler

import (
	"fmt"
	"github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
	"os"
	"testing"
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
