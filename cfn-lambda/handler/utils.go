package handler

import (
	"context"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/logzio/firehose-logs/common"
)

const lambdaToTrigger = "log-group-events-lambda"

type LambdaClient struct {
	Function lambdaiface.LambdaAPI
}

func invokeLambdaSynchronously(ctx context.Context, payload []byte) (string, error) {
	client, err := createLambdaClient()
	if err != nil {
		sugLog.Error("Error creating lambda client: ", err.Error())
		return "", err
	}

	res, err := client.invokeLambda(ctx, lambdaToTrigger, "RequestResponse", payload)
	if err != nil {
		return "", err
	}
	return string(res.Payload), nil
}

func invokeLambdaAsynchronously(ctx context.Context, payload []byte) error {
	client, err := createLambdaClient()
	if err != nil {
		sugLog.Error("Error creating lambda client: ", err.Error())
		return err
	}

	_, err = client.invokeLambda(ctx, lambdaToTrigger, "Event", payload)
	return err
}

func (client *LambdaClient) invokeLambda(ctx context.Context, functionName, invocationType string, payload []byte) (*lambda.InvokeOutput, error) {
	sugLog.Debugf("Invoking lambda %s %s", functionName, invocationType)

	res, err := client.Function.InvokeWithContext(ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(functionName),
		InvocationType: aws.String(invocationType),
		Payload:        payload,
	})

	if err != nil {
		sugLog.Error("Error while invoking lambda: ", err.Error())
		return nil, err
	}
	return res, nil
}

func createLambdaClient() (*LambdaClient, error) {
	sess, err := common.GetSession()
	if err != nil {
		sugLog.Error("Error while creating session: ", err.Error())
		return nil, err
	}
	return &LambdaClient{Function: lambda.New(sess)}, nil
}
