package handler

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager/types"
	"github.com/logzio/firehose-logs/common"
	lp "github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"os"
	"testing"
	"time"
)

func stringPtr(s string) *string {
	/* helper function */
	return &s
}

func timePtr(t time.Time) *time.Time {
	/* helper function */
	return &t
}

type MockSecretManagerClient struct {
	mock.Mock
	SecretsManagerAPIInterface
}

func (m *MockSecretManagerClient) GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error) {
	args := m.Called(params)
	return args.Get(0).(*secretsmanager.GetSecretValueOutput), args.Error(1)

}

func (m *MockSecretManagerClient) ListSecretVersionIds(ctx context.Context, params *secretsmanager.ListSecretVersionIdsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretVersionIdsOutput, error) {
	args := m.Called(params)
	return args.Get(0).(*secretsmanager.ListSecretVersionIdsOutput), args.Error(1)
}

func setupSecretTest() (ctx context.Context, mockClient *MockSecretManagerClient) {
	/* Setup logger */
	sugLog = lp.GetSugaredLogger()

	return context.Background(), new(MockSecretManagerClient)
}

func TestGetSecretNameFromArn(t *testing.T) {
	err := os.Setenv(common.EnvAwsRegion, "us-east-1")
	if err != nil {
		return
	}
	err = os.Setenv(envAccountId, "486140753397")
	if err != nil {
		return
	}

	tests := []struct {
		name               string
		arn                string
		expectedSecretName string
	}{
		{
			name:               "camel case secret name",
			arn:                "arn:aws:secretsmanager:us-east-1:486140753397:secret:testSecretName-56y7ud",
			expectedSecretName: "testSecretName",
		},
		{
			name:               "kebab case secret name",
			arn:                "arn:aws:secretsmanager:us-east-1:486140753397:secret:random-name-56y7ud",
			expectedSecretName: "random-name",
		},
		{
			name:               "snake case secret name",
			arn:                "arn:aws:secretsmanager:us-east-1:486140753397:secret:random_name-56y7ud",
			expectedSecretName: "random_name",
		},
		{
			name:               "name with numbers",
			arn:                "arn:aws:secretsmanager:us-east-1:486140753397:secret:now1with2numbers345-56y7ud",
			expectedSecretName: "now1with2numbers345",
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			assert.Equal(t, test.expectedSecretName, getSecretNameFromArn(test.arn))
		})
	}
}

func TestGetOldSecretValue(t *testing.T) {
	ctx, mockClient := setupSecretTest()

	tests := []struct {
		name           string
		SecretString   string
		versions       []types.SecretVersionsListEntry
		expectedOutput []string
		expectedError  bool
	}{
		{
			name:         "valid secret with history",
			SecretString: `{"logzioCustomLogGroups": "g1, g2, g3"}`,
			versions: []types.SecretVersionsListEntry{
				{
					VersionId:   stringPtr("v2"),
					CreatedDate: timePtr(time.Date(2023, 10, 8, 12, 0, 0, 0, time.UTC)),
				},
				{
					VersionId:   stringPtr("v1"),
					CreatedDate: timePtr(time.Date(2023, 5, 8, 12, 0, 0, 0, time.UTC)),
				},
			},
			expectedOutput: []string{"g1", "g2", "g3"},
			expectedError:  false,
		},
		{
			name:         "valid secret with no history",
			SecretString: "",
			versions: []types.SecretVersionsListEntry{
				{
					VersionId:   stringPtr("v1"),
					CreatedDate: timePtr(time.Date(2023, 5, 8, 12, 0, 0, 0, time.UTC)),
				},
			},
			expectedOutput: []string{},
			expectedError:  true,
		},
		{
			name:         "invalid last secret",
			SecretString: `{"someKey": "g1, g2, g3"}`,
			versions: []types.SecretVersionsListEntry{
				{
					VersionId:   stringPtr("v2"),
					CreatedDate: timePtr(time.Date(2023, 10, 8, 12, 0, 0, 0, time.UTC)),
				},
				{
					VersionId:   stringPtr("v1"),
					CreatedDate: timePtr(time.Date(2023, 5, 8, 12, 0, 0, 0, time.UTC)),
				},
			},
			expectedOutput: []string{},
			expectedError:  true,
		},
	}

	arn := "arn:aws:secretsmanager:us-east-1:486140753397:secret:testSecretName-56y7ud"
	secretName := "testSecretName"
	oldVer := "v1"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockClient.On("GetSecretValue", mock.Anything).Return(
				&secretsmanager.GetSecretValueOutput{
					ARN:          stringPtr(arn),
					Name:         stringPtr(secretName),
					VersionId:    stringPtr(oldVer),
					SecretString: stringPtr(test.SecretString),
				}, nil).Once()
			mockClient.On("ListSecretVersionIds", mock.Anything).Return(
				&secretsmanager.ListSecretVersionIdsOutput{
					ARN:      stringPtr(arn),
					Name:     stringPtr(secretName),
					Versions: test.versions,
				}, nil).Once()

			secretClient := &SecretManagerClient{Client: mockClient}
			result, err := secretClient.getOldSecretValue(ctx, secretName)

			assert.Equal(t, test.expectedOutput, result)
			if test.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestGetPreviousSecretVersion(t *testing.T) {
	ctx, mockClient := setupSecretTest()

	tests := []struct {
		name           string
		versions       []types.SecretVersionsListEntry
		expectedOutput *string
		expectedError  bool
	}{
		{
			name: "secret with history",
			versions: []types.SecretVersionsListEntry{
				{
					VersionId:   stringPtr("v2"),
					CreatedDate: timePtr(time.Date(2023, 10, 8, 12, 0, 0, 0, time.UTC)),
				},
				{
					VersionId:   stringPtr("v1"),
					CreatedDate: timePtr(time.Date(2023, 5, 8, 12, 0, 0, 0, time.UTC)),
				},
			},
			expectedOutput: stringPtr("v1"),
			expectedError:  false,
		},
		{
			name: "secret with no history",
			versions: []types.SecretVersionsListEntry{
				{
					VersionId:   stringPtr("v1"),
					CreatedDate: timePtr(time.Date(2023, 5, 8, 12, 0, 0, 0, time.UTC)),
				},
			},
			expectedOutput: nil,
			expectedError:  true,
		},
	}

	arn := "arn:aws:secretsmanager:us-east-1:486140753397:secret:testSecretName-56y7ud"
	secretName := "testSecretName"

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			mockClient.On("ListSecretVersionIds", mock.Anything).Return(
				&secretsmanager.ListSecretVersionIdsOutput{
					ARN:      stringPtr(arn),
					Name:     stringPtr(secretName),
					Versions: test.versions,
				}, nil).Once()
			secretClient := &SecretManagerClient{Client: mockClient}
			result, err := secretClient.getPreviousSecretVersion(ctx, secretName)

			assert.Equal(t, test.expectedOutput, result)
			if test.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}

func TestExtractCustomGroupsFromSecret(t *testing.T) {
	tests := []struct {
		name           string
		secretId       string
		result         string
		expectedOutput string
		expectedError  bool
	}{
		{
			name:           "valid secret",
			secretId:       "testSecret",
			result:         `{"logzioCustomLogGroups": "g1, g2, g3"}`,
			expectedOutput: "g1, g2, g3",
			expectedError:  false,
		},
		{
			name:           "empty secret",
			secretId:       "testSecret",
			result:         ``,
			expectedOutput: "",
			expectedError:  true,
		},
		{
			name:           "missing `logzioSecretKeyName` as key",
			secretId:       "testSecret",
			result:         `{"someKey": "g1, g2"}`,
			expectedOutput: "",
			expectedError:  true,
		},
		{
			name:           "edge case, invalid json",
			secretId:       "testSecret",
			result:         `{""someKey": "g1, g2"}`,
			expectedOutput: "",
			expectedError:  true,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			result, err := extractCustomGroupsFromSecret(test.secretId, test.result)
			assert.Equal(t, test.expectedOutput, result)
			if test.expectedError {
				assert.NotNil(t, err)
			} else {
				assert.Nil(t, err)
			}
		})
	}
}
