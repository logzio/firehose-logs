package handler

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/secretsmanager"
	"github.com/aws/aws-secretsmanager-caching-go/secretcache"
	"github.com/logzio/firehose-logs/common"
	"os"
	"regexp"
	"sort"
)

// SecretsManagerAPIInterface AWS SDK v2 doesn't provide an interface for each service client like v1
type SecretsManagerAPIInterface interface {
	GetSecretValue(ctx context.Context, params *secretsmanager.GetSecretValueInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.GetSecretValueOutput, error)
	ListSecretVersionIds(ctx context.Context, params *secretsmanager.ListSecretVersionIdsInput, optFns ...func(*secretsmanager.Options)) (*secretsmanager.ListSecretVersionIdsOutput, error)
}

type SecretCacheInterface interface {
	GetSecretString(string) (string, error)
}

type SecretManagerClient struct {
	Client SecretsManagerAPIInterface
}

type SecretCacheClient struct {
	Client SecretCacheInterface
}

func getSecretManagerClient(ctx context.Context) (*SecretManagerClient, error) {
	awsConf, err := config.LoadDefaultConfig(ctx, config.WithRegion(os.Getenv(common.EnvAwsRegion)))
	if err != nil {
		sugLog.Error("Failed to setup connection to get older custom log groups secret values.")
		return nil, err
	}
	return &SecretManagerClient{Client: secretsmanager.NewFromConfig(awsConf)}, nil
}

func getSecretCacheClient() (*SecretCacheClient, error) {
	secretCache, err := secretcache.New()

	return &SecretCacheClient{Client: secretCache}, err
}

// updateSecretCustomLogGroups updates the custom log groups to monitor based on comparing the old secret value to the new one
func updateSecretCustomLogGroups(ctx context.Context, secretId string) error {
	svc, err := getSecretManagerClient(ctx)
	if err != nil {
		return err
	}

	oldSecretValue, err := svc.getOldSecretValue(ctx, secretId)
	if err != nil {
		sugLog.Error("Failed to get the old custom log group secret version's value.")
		return err
	}

	newSecretValue, err := getCustomLogGroups("true", secretId)
	if err != nil {
		sugLog.Error("Failed to get the new custom log group from secret")
		return err
	}

	customGroupsToAdd, customGroupsToRemove := findDifferences(oldSecretValue, newSecretValue)

	cwLogClient, err := getCloudWatchLogsClient()
	if err != nil {
		sugLog.Error("Failed to get the old custom log group secret version's value.")
		return err
	}

	if err := cwLogClient.updateSubscriptionFilters([]string{}, []string{}, customGroupsToAdd, customGroupsToRemove); err != nil {
		return err
	}
	return nil
}

// getSecretNameFromArn extracts a secret name from the given secret ARN
func getSecretNameFromArn(secretArn string) string {
	var secretName string

	getSecretName := regexp.MustCompile(fmt.Sprintf(`^arn:aws:secretsmanager:%s:%s:secret:(?P<secretName>\S+)-`, os.Getenv(common.EnvAwsRegion), os.Getenv(envAccountId)))
	match := getSecretName.FindStringSubmatch(secretArn)

	for i, key := range getSecretName.SubexpNames() {
		if key == "secretName" && len(match) > i {
			secretName = match[i]
			break
		}
	}
	sugLog.Debugf("Found secret name %s, from secret ARN %s", secretName, secretArn)
	return secretName
}

// getOldSecretValue gets custom log groups value from the previous secret version
func (svc *SecretManagerClient) getOldSecretValue(ctx context.Context, secretId string) ([]string, error) {
	// get the old version id
	oldVersionId, err := svc.getPreviousSecretVersion(ctx, secretId)
	if err != nil {
		sugLog.Error("Failed to get the older custom log group secret version.")
		return []string{}, err
	}

	// get the old version value
	getOldSecretValueInput := &secretsmanager.GetSecretValueInput{
		SecretId:  &secretId,
		VersionId: oldVersionId,
	}

	oldSecret, err := svc.Client.GetSecretValue(ctx, getOldSecretValueInput)
	if err != nil {
		sugLog.Error("Failed to get the old value of the secret")
		return []string{}, err
	}
	oldSecretValueJson := oldSecret.SecretString

	oldSecretValue, err := extractCustomGroupsFromSecret(secretId, *oldSecretValueJson)
	if err != nil {
		sugLog.Error("Failed to get the old value of the custom log groups from secret")
		return []string{}, err
	}

	return convertStrToArr(oldSecretValue), nil

}

// getPreviousSecretVersion returns the previous version id of the given secret
func (svc *SecretManagerClient) getPreviousSecretVersion(ctx context.Context, secretId string) (*string, error) {
	var previousVersionId *string

	listSecretVersionsInput := &secretsmanager.ListSecretVersionIdsInput{
		SecretId: &secretId,
	}

	secretInfo, err := svc.Client.ListSecretVersionIds(ctx, listSecretVersionsInput)
	if err != nil {
		sugLog.Error("Failed to list secret versions")
		return nil, err
	}

	// Sort the versions based on created date
	sort.Slice(secretInfo.Versions, func(i, j int) bool {
		return secretInfo.Versions[i].CreatedDate.After(*secretInfo.Versions[j].CreatedDate)
	})

	if len(secretInfo.Versions) > 1 {
		previousVersionId = secretInfo.Versions[1].VersionId
	} else {
		return nil, fmt.Errorf("secret %s doesn't have an older version", secretId)
	}

	return previousVersionId, nil
}

// extractCustomGroupsFromSecret extracts the custom log groups to monitor from the given secret value
func extractCustomGroupsFromSecret(secretId, result string) (string, error) {
	var secretValues map[string]string
	err := json.Unmarshal([]byte(result), &secretValues)
	if err != nil {
		return "", err
	}

	customLogGroups, ok := secretValues[logzioSecretKeyName]
	if !ok {
		return "", fmt.Errorf("did not find logzioCustomLogGroups key in the secret %s", secretId)
	}
	return customLogGroups, nil
}
