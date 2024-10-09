package main

import (
	"github.com/aws/aws-lambda-go/lambda"
	handler "github.com/logzio/firehose-logs/log-group-events-lambda/handler"
)

func main() {
	lambda.Start(handler.HandleRequest)
}
