package handler

import (
	"github.com/logzio/firehose-logs/common"
	"os"
)

type Config struct {
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
	err := validateRequired()
	if err != nil {
		sugLog.Error("Error while validating required environment variables: ", err)
		return nil
	}

	return &Config{
		destinationArn:       os.Getenv(envFirehoseArn),
		roleArn:              os.Getenv(envPutSubscriptionFilterRole),
		accountId:            os.Getenv(envAccountId),
		region:               os.Getenv(common.EnvAwsRegion),
		thisFunctionLogGroup: lambdaPrefix + os.Getenv(envFunctionName),
		thisFunctionName:     os.Getenv(envFunctionName),
		customGroupsValue:    os.Getenv(common.EnvCustomGroups),
		servicesValue:        os.Getenv(common.EnvServices),
	}
}
