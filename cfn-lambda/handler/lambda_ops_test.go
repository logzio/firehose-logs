package handler

import (
	"context"
	"encoding/json"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/request"
	"github.com/aws/aws-sdk-go/service/lambda"
	"github.com/aws/aws-sdk-go/service/lambda/lambdaiface"
	"github.com/logzio/firehose-logs/common"
	lp "github.com/logzio/firehose-logs/logger"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type mockLambdaClient struct {
	mock.Mock
	lambdaiface.LambdaAPI
}

func (m *mockLambdaClient) InvokeWithContext(ctx aws.Context, input *lambda.InvokeInput, opts ...request.Option) (*lambda.InvokeOutput, error) {
	args := m.Called(ctx, input)
	return &lambda.InvokeOutput{
		StatusCode:    aws.Int64(200),
		FunctionError: nil,
	}, args.Error(1)
}

func setupLambdaTest(eventType common.ActionType) (ctx context.Context, payload []byte) {
	/* Setup mock context and event */
	mockEvent := common.NewSubscriptionFilterEvent(common.RequestParameters{
		Action:      eventType,
		NewServices: "service1, service2",
		OldServices: "service1, service3",
		NewCustom:   "log-group1, log-group2",
		OldCustom:   "log-group1",
		NewIsSecret: "false",
		OldIsSecret: "false",
	})
	mockEventPayload, _ := json.Marshal(mockEvent)

	ctx = context.Background()

	/* Setup logger */
	sugLog = lp.GetSugaredLogger()

	return ctx, mockEventPayload
}

func TestInvokeLambdaAsynchronously(t *testing.T) {
	ctx, mockEvent := setupLambdaTest(common.AddSF)

	/* mock the lambda client */
	mockClient := new(mockLambdaClient)
	mockClient.On("InvokeWithContext", ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(lambdaToTrigger),
		InvocationType: aws.String("Event"),
		Payload:        mockEvent,
	}).Return(&lambda.InvokeOutput{
		StatusCode:    aws.Int64(200),
		FunctionError: nil,
	}, nil)

	/* test trigger */
	lambdaClient := &LambdaClient{Function: mockClient}
	_, err := lambdaClient.invokeLambda(ctx, lambdaToTrigger, "Event", mockEvent)

	assert.NoError(t, err)
	mockClient.AssertExpectations(t)
}

func TestInvokeLambdaSynchronously(t *testing.T) {
	ctx, mockEvent := setupLambdaTest(common.AddSF)

	/* mock the lambda client */
	mockClient := new(mockLambdaClient)
	mockClient.On("InvokeWithContext", ctx, &lambda.InvokeInput{
		FunctionName:   aws.String(lambdaToTrigger),
		InvocationType: aws.String("RequestResponse"),
		Payload:        mockEvent,
	}).Return(&lambda.InvokeOutput{
		StatusCode:    aws.Int64(200),
		FunctionError: nil,
	}, nil)

	/* test trigger */
	lambdaClient := &LambdaClient{Function: mockClient}
	res, err := lambdaClient.invokeLambda(ctx, lambdaToTrigger, "RequestResponse", mockEvent)

	assert.NoError(t, err)
	assert.Equal(t, &lambda.InvokeOutput{
		StatusCode:    aws.Int64(200),
		FunctionError: nil,
	}, res)
	mockClient.AssertExpectations(t)
}
