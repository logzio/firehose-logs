package handler

import (
	"fmt"
	"github.com/logzio/firehose-logs/common"
	lp "github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"sort"
	"testing"
)

type MockSecretCacheClient struct {
	mock.Mock
	SecretCacheInterface
}

func (m *MockSecretCacheClient) GetSecretString(secretName string) (string, error) {
	if secretName == "errorSecret" {
		return "", fmt.Errorf("an error occurred")
	}
	args := m.Called(secretName)
	return args.String(0), args.Error(1)
}

func setupLGTest() (cwClient *CloudWatchLogsClient, secretCacheClient *MockSecretCacheClient) {
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

	/* Setup config */
	envConfig = NewConfig()

	/* Setup logger */
	sugLog = lp.GetSugaredLogger()

	/* Setup mock */
	mockCwClient := new(MockCloudWatchLogsClient)
	return &CloudWatchLogsClient{Client: mockCwClient}, new(MockSecretCacheClient)
}

func TestGetServices(t *testing.T) {
	/* No services */
	err := os.Unsetenv(common.EnvServices)
	if err != nil {
		return
	}

	setupLGTest()

	result := getServices()
	assert.Nil(t, result)

	/* Has services */
	err = os.Setenv(common.EnvServices, "rds, cloudwatch, custom")
	if err != nil {
		return
	}
	setupLGTest()

	result = getServices()
	assert.Equal(t, []string{"rds", "cloudwatch", "custom"}, result)
}

func TestGetServicesLogGroups(t *testing.T) {
	cwClient, _ := setupLGTest()

	tests := []struct {
		name              string
		services          []string
		expectedLogGroups []string
	}{
		{
			name:              "valid services",
			services:          []string{"cloudtrail", "apigateway"},
			expectedLogGroups: []string{"/aws/apigateway/g1", "/random/log/group"},
		},
		{
			name:              "invalid services",
			services:          []string{"svc1", "svc2"},
			expectedLogGroups: []string{},
		},
		{
			name:              "empty services",
			services:          []string{},
			expectedLogGroups: []string{},
		},
		{
			name:              "don't monitor this function's log group",
			services:          []string{"lambda"},
			expectedLogGroups: []string{"/aws/lambda/g1"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result := getServicesLogGroups(test.services, cwClient)
			sort.Strings(result)
			assert.Equal(t, test.expectedLogGroups, result)
		})
	}
}

func TestGetCustomLogGroups(t *testing.T) {
	setupLGTest()
	tests := []struct {
		name                  string
		secretEnabled         string
		customLogGroupsPrmVal string
		expectedGroups        []string
	}{
		{
			name:                  "no log groups",
			secretEnabled:         "false",
			customLogGroupsPrmVal: "",
			expectedGroups:        []string{},
		},
		{
			name:                  "multiple log groups",
			secretEnabled:         "false",
			customLogGroupsPrmVal: "g1, g2, g3",
			expectedGroups:        []string{"g1", "g2", "g3"},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := getCustomLogGroups(test.secretEnabled, test.customLogGroupsPrmVal)
			sort.Strings(result)
			assert.Equal(t, test.expectedGroups, result)
			assert.Nil(t, err)
		})
	}
}

func TestGetCustomLogGroupsFromSecret(t *testing.T) {
	/* set needed env variables */
	err := os.Setenv(common.EnvAwsRegion, "us-west-2")
	if err != nil {
		return
	}

	err = os.Setenv(envAccountId, "123456789012")
	if err != nil {
		return
	}

	/* get mocks */
	_, mockSecretCacheClient := setupLGTest()

	/* tests */
	tests := []struct {
		name          string
		secretArn     string
		secretData    string
		expectedData  []string
		expectedError bool
	}{
		{
			name:          "no log groups",
			secretArn:     "arn:aws:secretsmanager:us-west-2:123456789012:secret:my-secret",
			secretData:    `{"logzioCustomLogGroups": ""}`,
			expectedData:  nil,
			expectedError: false,
		},
		{
			name:          "multiple log groups",
			secretArn:     "arn:aws:secretsmanager:us-west-2:123456789012:secret:my-secret",
			secretData:    `{"logzioCustomLogGroups": "g1, g2, g3"}`,
			expectedData:  []string{"g1", "g2", "g3"},
			expectedError: false,
		},
		{
			name:          "error in getting secret",
			secretArn:     "arn:aws:secretsmanager:us-west-2:123456789012:errorSecret",
			secretData:    `{"someKey": "value"}`,
			expectedData:  nil,
			expectedError: true,
		},
		{
			name:          "invalid secret value",
			secretArn:     "arn:aws:secretsmanager:us-west-2:123456789012:secret:my-secret",
			secretData:    `{"somKey": "g1, g2"}`,
			expectedData:  nil,
			expectedError: true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {

			mockSecretCacheClient.On("GetSecretString", mock.Anything).Return(test.secretData, nil).Once()
			secretCacheClient := &SecretCacheClient{Client: mockSecretCacheClient}

			result, err := getCustomLogGroupsFromSecret(test.secretArn, secretCacheClient, nil)
			sort.Strings(result)
			assert.Equal(t, test.expectedData, result)

			if test.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetCustomLogGroupsFromParam(t *testing.T) {
	/* Note:
	TestGetCustomLogGroups and TestGetCustomLogGroupsFromSecret tests this function for non prefix log groups.
	This test will focus on testing the prefix via wildcard support capability.
	*/
	cwClient, _ := setupLGTest()

	tests := []struct {
		name           string
		logGroups      []string
		expectedGroups []string
		expectedError  bool
	}{
		{
			name:           "all wildcards",
			logGroups:      []string{"/log/group1/*", "/log/group2/*"},
			expectedGroups: []string{"/log/group1/a", "/log/group1/b", "/log/group2/a"},
			expectedError:  false,
		},
		{
			name:           "some wildcards, some not wildcards",
			logGroups:      []string{"/log/group1/*", "g1", "g2"},
			expectedGroups: []string{"/log/group1/a", "/log/group1/b", "g1", "g2"},
			expectedError:  false,
		},
		{
			name:           "no sub groups for prefix",
			logGroups:      []string{"/aws/codebuild/*"},
			expectedGroups: []string{},
			expectedError:  false,
		},
		{
			name:           "one error",
			logGroups:      []string{"/aws/error/test/*", "/log/group2/*", "g3"},
			expectedGroups: []string{"/log/group2/a", "g3"},
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := getCustomLogGroupsFromParam(test.logGroups, cwClient)
			sort.Strings(result)
			assert.Equal(t, test.expectedGroups, result)

			if test.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
