package main

import (
	"github.com/aws/aws-lambda-go/cfn"
	"github.com/aws/aws-lambda-go/lambda"
	handler "github.com/logzio/firehose-logs/cfn-lambda/handler"
)

func main() {
	lambda.Start(cfn.LambdaWrap(handler.HandleRequest))
}
