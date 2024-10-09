package handler

const (
	envFunctionName              = "AWS_LAMBDA_FUNCTION_NAME" // reserved env
	envAccountId                 = "ACCOUNT_ID"
	envFirehoseArn               = "FIREHOSE_ARN"
	envAwsPartition              = "AWS_PARTITION"
	envPutSubscriptionFilterRole = "PUT_SF_ROLE"

	logzioSecretKeyName    = "logzioCustomLogGroups"
	valuesSeparator        = ","
	emptyString            = ""
	lambdaPrefix           = "/aws/lambda/"
	subscriptionFilterName = "logzio_firehose"
)
