package handler

import (
	"fmt"
	"github.com/logzio/firehose-logs/common"
	"os"
)

type Config struct {
	awsPartition         string
	destinationArn       string
	roleArn              string
	accountId            string
	region               string
	thisFunctionLogGroup string
	thisFunctionName     string
	customGroupsValue    string
	servicesValue        string
}

func NewConfig() *Config {
	c := Config{
		awsPartition:         os.Getenv(envAwsPartition),
		destinationArn:       os.Getenv(envFirehoseArn),
		roleArn:              os.Getenv(envPutSubscriptionFilterRole),
		accountId:            os.Getenv(envAccountId),
		region:               os.Getenv(common.EnvAwsRegion),
		thisFunctionLogGroup: lambdaPrefix + os.Getenv(envFunctionName),
		thisFunctionName:     os.Getenv(envFunctionName),
		customGroupsValue:    os.Getenv(common.EnvCustomGroups),
		servicesValue:        os.Getenv(common.EnvServices),
	}

	err := c.validateRequired()
	if err != nil {
		sugLog.Error("Error while validating required environment variables: ", err)
		return nil
	}
	return &c
}

func (c *Config) validateRequired() error {
	if c.destinationArn == emptyString {
		return fmt.Errorf("destination ARN must be set")
	}

	if c.accountId == emptyString {
		return fmt.Errorf("account id must be set")
	}

	if c.awsPartition == emptyString {
		return fmt.Errorf("aws partition must be set")
	}

	return nil
}
