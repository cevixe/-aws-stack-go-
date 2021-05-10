package main

import (
	"context"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cevixe/aws-sdk-go/aws/impl"
	"github.com/cevixe/aws-sdk-go/aws/integration/dynamodb"
	"github.com/cevixe/aws-sdk-go/aws/model"
	"github.com/cevixe/aws-sdk-go/aws/runtime"
	"sync"
)

func handler(ctx context.Context, input events.DynamoDBEvent) error {

	awsContext := ctx.Value(impl.AwsContext).(*impl.Context)
	wg := &sync.WaitGroup{}
	wg.Add(len(input.Records))
	for _, record := range input.Records {
		go publishEvent(ctx, awsContext.AwsEventBus, record, wg)
	}
	wg.Wait()
	return nil
}

func publishEvent(ctx context.Context, bus model.AwsEventBus, record events.DynamoDBEventRecord, wg *sync.WaitGroup) {
	if record.Change.NewImage == nil {
		wg.Done()
	}
	eventRecord := &model.AwsEventRecord{}
	dynamodb.UnmarshallDynamodbStreamItem(record.Change.NewImage, eventRecord)
	bus.PublishEvent(ctx, eventRecord)
	wg.Done()
}

func main() {
	ctx := runtime.NewContext()
	lambda.StartWithContext(ctx, handler)
}
