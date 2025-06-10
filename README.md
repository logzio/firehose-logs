# Shipping logs from Cloudwatch into Logz.io with Firehose Delivery Stream

This project deploys instrumentation that allows shipping Cloudwatch logs to Logz.io, with a Firehose Delivery Stream.

## Overview

This project will use a Cloudformation template to create a Stack that deploys:
* Firehose Delivery Stream with Logz.io as the stream's destination.
* Lambda function that adds Subscription Filters to Cloudwatch Log Groups, as defined by user's input.
* Roles, log groups, and other resources that are necessary for this instrumentation.

## Instructions

To deploy this project, click the button that matches the region you wish to deploy your Stack to:

| Region            | Deployment                                                                                                                                                                                                                                                                                                               |
|-------------------|--------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `us-east-1`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-us-east-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) | 
| `us-east-2`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-east-2#/stacks/create/review?templateURL=https://logzio-aws-integrations-us-east-2.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) | 
| `us-west-1`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-us-west-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) | 
| `us-west-2`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=us-west-2#/stacks/create/review?templateURL=https://logzio-aws-integrations-us-west-2.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) | 
| `eu-central-1`    | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-central-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-eu-central-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `eu-central-2`    | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-central-2#/stacks/create/review?templateURL=https://logzio-aws-integrations-eu-central-2.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `eu-north-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-north-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-eu-north-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `eu-west-1`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-eu-west-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `eu-west-2`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-2#/stacks/create/review?templateURL=https://logzio-aws-integrations-eu-west-2.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `eu-west-3`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-west-3#/stacks/create/review?templateURL=https://logzio-aws-integrations-eu-west-3.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `eu-south-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-south-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-eu-south-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `eu-south-2`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=eu-south-2#/stacks/create/review?templateURL=https://logzio-aws-integrations-eu-south-2.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `sa-east-1`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=sa-east-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-sa-east-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-northeast-1`  | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-northeast-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-northeast-2`  | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-2#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-northeast-2.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-northeast-3`  | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-northeast-3#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-northeast-3.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-south-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-south-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-south-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-south-2`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-south-2#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-south-2.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-southeast-1`  | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-southeast-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-southeast-2`  | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-2#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-southeast-2.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-southeast-3`  | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-3#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-southeast-3.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-southeast-4`  | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-southeast-4#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-southeast-4.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ap-east-1`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ap-east-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-ap-east-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ca-central-1`    | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ca-central-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-ca-central-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `ca-west-1`       | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=ca-west-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-ca-west-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `af-south-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=af-south-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-af-south-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `me-south-1`      | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=me-south-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-me-south-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `me-central-1`    | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=me-central-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-me-central-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |
| `il-central-1`    | [![Deploy to AWS](https://dytvr9ot2sszz.cloudfront.net/logz-docs/lights/LightS-button.png)](https://console.aws.amazon.com/cloudformation/home?region=il-central-1#/stacks/create/review?templateURL=https://logzio-aws-integrations-il-central-1.s3.amazonaws.com/firehose-logs/0.3.2/sam-template.yaml&stackName=logzio-firehose) |

### 1. Specify stack details

Specify the stack details as per the table below, check the checkboxes and select **Create stack**.

| Parameter                                  | Description                                                                                                                                                                                                                                                                                                                                                                                                                      | Required/Default  |
|--------------------------------------------|----------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|-------------------|
| `logzioToken`                              | The [token](https://app.logz.io/#/dashboard/settings/general) of the account you want to ship logs to.                                                                                                                                                                                                                                                                                                                           | **Required**      |
| `logzioListener`                           | Listener host.                                                                                                                                                                                                                                                                                                                                                                                                                   | **Required**      |
| `logzioType`                               | The log type you'll use with this Lambda. This can be a [built-in log type](https://docs.logz.io/user-guide/log-shipping/built-in-log-types.html), or a custom log type.                                                                                                                                                                                                                                                         | `logzio_firehose` |
| `services`                                 | A comma-seperated list of services you want to collect logs from. Supported options are: `apigateway`, `rds`, `cloudhsm`, `cloudtrail`, `codebuild`, `connect`, `elasticbeanstalk`, `ecs`, `eks`, `aws-glue`, `aws-iot`, `lambda`, `macie`, `amazon-mq`, `batch`                                                                                                                                                                 | -                 |
| `customLogGroups`                          | A comma-separated list of custom log groups to collect logs from, or the ARN of the Secret parameter ([explanation below](#custom-log-group-list-exceeds-4096-characters-limit)) storing the log groups list if it exceeds 4096 characters. **Note**: You can also specify a prefix of the log group names by using a wildcard at the end (e.g., `prefix*`). This will match all log groups that start with the specified prefix | -                 |
| `useCustomLogGroupsFromSecret`             | If you want to provide list of `customLogGroups` which exceeds 4096 characters, set to `true` and configure your customLogGroups as [defined below](#custom-log-group-list-exceeds-4096-characters-limit).                                                                                                                                                                                                                       | `false`           |
| `triggerLambdaTimeout`                     | The amount of seconds that Lambda allows a function to run before stopping it, for the trigger function.                                                                                                                                                                                                                                                                                                                         | `60`              |
| `triggerLambdaMemory`                      | Trigger function's allocated CPU proportional to the memory configured, in MB.                                                                                                                                                                                                                                                                                                                                                   | `512`             |
| `triggerLambdaLogLevel`                    | Log level for the Lambda function. Can be one of: `debug`, `info`, `warn`, `error`, `fatal`, `panic`                                                                                                                                                                                                                                                                                                                             | `info`            |
| `httpEndpointDestinationIntervalInSeconds` | The length of time, in seconds, that Kinesis Data Firehose buffers incoming data before delivering it to the destination                                                                                                                                                                                                                                                                                                         | `60`              |
| `httpEndpointDestinationSizeInMBs`         | The size of the buffer, in MBs, that Kinesis Data Firehose uses for incoming data before delivering it to the destination                                                                                                                                                                                                                                                                                                        | `5`               |


> #### ⚠️ Important note ⚠️
> AWS limits every log group to have up to 2 subscription filters. If your chosen log group already has 2 subscription filters, the trigger function won't be able to add another one.

<details>
  <summary>
    <h4>Guide if customLogGroups list exceeds 4096 characters limit</h4>
  </summary>

#### Custom Log Group list exceeds 4096 characters limit
If your `customLogGroups` list exceeds the 4096 characters limit, follow the below steps:

1. Open AWS [Secret Manager](https://console.aws.amazon.com/secretsmanager/)
2. Click `Store a new secret`
   - Choose `Other type of secret`
   - For `key` use `logzioCustomLogGroups`
   - In `value` store your comma-separated custom log groups list
   - Name your secret, for example as `LogzioCustomLogGroups`
   - Copy the new secret's ARN
3. In your stack, Set: 
   - `customLogGroups` to your secret ARN that you copied in step 2
   - `useCustomLogGroupsFromSecret` to `true`

</details>

### 2. Send logs

Give the stack a few minutes to be deployed.

Once new logs are added to your chosen log group, they will be sent to your Logz.io account.

> ##### ⚠️ Important note ⚠️
> If you've used the `services` field, you'll have to **wait 6 minutes** before creating new log groups for your chosen services. This is due to cold start and custom resource invocation, that can cause the Lambda to behave unexpectedly.

### Changelog:
- **0.3.3**:
  - Fix timing issue to make sure bucket is created before the delivery stream
  - Fix issue where EventBridge trigger for log group creation was not created
- **0.3.2**:
  - Fix issue where EventBridge trigger for log group creation was not created when using only `customLogGroups`.
- **0.3.1**:
    - Support deploying multiple stacks within the same AWS account
    - Resolve bug with update mechanism
- **0.3.0**: 
  - Support prefixes in `customLogGroups` via wildcard
  - Upgrade go `1.19` >> `1.22`
  - Parallelized subscription filter updates to improve performance
- **0.2.1**: Add support for `aws-batch` service.
- **0.2.0**: Option to provide `customLogGroups` exceeding 4KB.
- **0.1.0**:
  Introduced the ability to directly update service and custom log parameters within the stack.
- **0.0.2**: Fix for RDS service - look for prefix `/aws/rds/`
- **0.0.1**: Initial release.
