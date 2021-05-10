package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cevixe/aws-sdk-go/aws/impl"
	"github.com/cevixe/aws-sdk-go/aws/integration/dynamodb"
	"github.com/cevixe/aws-sdk-go/aws/model"
	"github.com/cevixe/aws-sdk-go/aws/runtime"
	"github.com/cevixe/core-sdk-go/core"
	"strconv"
)

func handler(ctx context.Context, input events.DynamoDBEvent) error {

	awsContext := ctx.Value(impl.AwsContext).(*impl.Context)

	eventRecords := make(map[string]*model.AwsEventRecord)
	for _, item := range input.Records {
		if item.Change.NewImage == nil {
			continue
		}
		eventRecord := &model.AwsEventRecord{}
		dynamodb.UnmarshallDynamodbStreamItem(item.Change.NewImage, eventRecord)
		if *eventRecord.EventClass == string(core.DomainEvent) {
			if eventRecord.Reference != nil && eventRecord.EventData == nil {
				awsContext.AwsObjectStore.GetObject(ctx, eventRecord.Reference, eventRecord)
			}
			eventRecords[*eventRecord.EventSource] = eventRecord
		}
	}

	stateRecords := make([]*model.AwsStateRecord, 0, len(eventRecords))
	for _, value := range eventRecords {
		version, err := strconv.ParseUint(*value.EventID, 10, 64)
		if err != nil {
			panic(fmt.Errorf("cannot unmarshall event id\n%v", err))
		}
		stateRecord := &model.AwsStateRecord{
			ID:        *value.EntityID,
			Type:      *value.EntityType,
			Version:   version,
			State:     *value.EntityState,
			UpdatedAt: *value.EventTime,
			UpdatedBy: *value.EventAuthor,
			CreatedAt: *value.EntityCreatedAt,
			CreatedBy: *value.EntityCreatedBy,
		}
		stateRecords = append(stateRecords, stateRecord)
	}

	if len(stateRecords) > 0 {
		awsContext.AwsStateStore.UpdateStates(ctx, stateRecords)
	}
	return nil
}

func main() {
	ctx := runtime.NewContext()
	lambda.StartWithContext(ctx, handler)
}
