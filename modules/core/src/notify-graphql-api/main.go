package main

import (
	"context"
	"fmt"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/cevixe/aws-sdk-go/aws/impl"
	"github.com/cevixe/aws-sdk-go/aws/integration/sqs"
	"github.com/cevixe/aws-sdk-go/aws/model"
	"github.com/cevixe/aws-sdk-go/aws/runtime"
	"github.com/cevixe/aws-sdk-go/aws/util"
	"github.com/cevixe/core-sdk-go/core"
	"reflect"
	"sync"
)

const gqlQuery string = `
mutation(
	$Input: AWSJSON!
) {
	publishMessage(
		input: $Input
	) {
		transaction
		eventType
		eventAuthor
		eventJson
        event {
          ...fullEvent
        }
		entityId
		entityType
		entityOwner
		entityJson
        entity {
          ...fullEntity
        }
	}
}
`

const eventStandardFields = `
	__typename
	_id
	_type
	_time
	_author
	_sourceId
	_sourceType
	_sourceVersion
	_transaction
`

const entityStandardFields = `
	__typename
    _id
    _type
    _version
    _createdAt
    _createdBy
    _updatedAt
    _updatedBy
`

func objectGqlFields(obj interface{}) string {
	var gql string
	objMap := obj.(*map[string]interface{})
	for key, value := range *objMap {
		if reflect.ValueOf(value).Kind() == reflect.Map {
			gql = gql + key + " {\n" + objectGqlFields(value) + "}\n"
		} else {
			gql = gql + key + "\n"
		}
	}
	return gql
}

func generateEventFragment(typ string, object interface{}) string {
	return "fragment fullEvent on Event {\n" +
		"...on " + typ + " {\n" +
		eventStandardFields +
		objectGqlFields(object) +
		"}\n" +
		"}\n"
}

func generateEntityFragment(typ string, object interface{}) string {
	return "fragment fullEntity on Entity {\n" +
		"...on " + typ + " {\n" +
		entityStandardFields +
		objectGqlFields(object) +
		"}\n" +
		"}\n"
}

func generateRequest(ctx context.Context, event *model.AwsEventRecord) *model.AwsGraphqlRequest {

	awsContext := ctx.Value(impl.AwsContext).(*impl.Context)
	if event.Reference != nil && event.EventData == nil {
		awsContext.AwsObjectStore.GetObject(ctx, event.Reference, event)
	}

	eventFragment := generateEventFragment(*event.EventType, event.EventData)
	entityFragment := generateEntityFragment(*event.EntityType, event.EventData)
	fullGqlQuery := gqlQuery + "\n" + eventFragment + "\n" + entityFragment

	return &model.AwsGraphqlRequest{
		Query: fullGqlQuery,
		Variables: map[string]interface{}{
			"Input": util.MarshalJsonString(map[string]interface{}{
				"transaction": event.Transaction,
				"eventType":   event.EventType,
				"eventAuthor": event.EventAuthor,
				"eventJson":   event.EventData,
				"entityId":    event.EntityID,
				"entityType":  event.EntityType,
				"entityOwner": event.EntityCreatedBy,
				"entityJson":  event.EntityState,
			}),
		},
	}
}

func handler(ctx context.Context, input events.SQSEvent) error {

	eventRecords := make([]*model.AwsEventRecord, 0, len(input.Records))
	sqs.UnmarshallSQSEvent(input, &eventRecords)

	wg := &sync.WaitGroup{}
	wg.Add(len(eventRecords))
	for _, item := range eventRecords {
		go asyncGraphqlCall(ctx, item, wg)
	}
	wg.Wait()
	return nil
}

func asyncGraphqlCall(ctx context.Context, event *model.AwsEventRecord, wg *sync.WaitGroup) {

	if *event.EventClass != string(core.DomainEvent) {
		wg.Done()
		return
	} else {
		request := generateRequest(ctx, event)
		fmt.Println("request:\n" + util.MarshalJsonString(request))
		//awsContext := ctx.Value(impl.AwsContext).(*impl.Context)
		//response := awsContext.AwsGraphqlGateway.ExecuteGraphql(ctx, request)
		//if len(response.Errors) > 0 {
		//	panic(fmt.Errorf("invalid gql request:\n%v", response.Errors))
		//}
		wg.Done()
	}
}

func main() {
	ctx := runtime.NewContext()
	lambda.StartWithContext(ctx, handler)
}
