package main

import (
	"context"
	"encoding/json"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/sns"
)

var eventDynamodbTable string
var eventSNSTopicArn string
var ddbClient *dynamodb.Client
var snsClient *sns.Client

type EventItem struct {
	EventId   string `dynamodbav:"eventId" json:"eventId"`
	Action    string `dynamodbav:"action" json:"action"`
	UserId    string `dynamodbav:"userId" json:"userId"`
	CreatedAt string `dynamodbav:"createdAt" json:"createdAt"`
	ObjectId  string `dynamodbav:"objectId" json:"objectId"`
	BizId     string `dynamodbav:"bizId" json:"bizId"`
	ErrorMsg  string `dynamodbav:"errorMsg" json:"errorMsg"`
}

type TableBasics struct {
	DynamoDbClient *dynamodb.Client
	TableName      string
}

// ensure idempotency with a condition expression
func (m TableBasics) putItem(eventItem *EventItem) (err error) {
	item, err := attributevalue.MarshalMap(eventItem)
	if err != nil {
		return
	}
	_, err = m.DynamoDbClient.PutItem(context.TODO(), &dynamodb.PutItemInput{
		TableName:           aws.String(m.TableName),
		Item:                item,
		ConditionExpression: aws.String("attribute_not_exists(inspectedAt)"),
	})
	if err != nil {
		log.Printf("Couldn't add item to table. Here's why: %v\n", err)
	}

	return
}

// more example: https://github.com/awsdocs/aws-doc-sdk-examples/tree/main/gov2
func Init() {
	eventDynamodbTable = os.Getenv("TABLE_NAME")
	eventSNSTopicArn = os.Getenv("TOPIC_ARN")
	if len(eventDynamodbTable) == 0 || len(eventSNSTopicArn) == 0 {
		log.Fatalf("env TABLE_NAME:%s TOPIC_ARN:%s is empty", eventDynamodbTable, eventSNSTopicArn)
	}

	cfg, err := config.LoadDefaultConfig(context.TODO(), config.WithRegion("us-east-1"))
	if err != nil {
		log.Fatalf("unable to load SDK config, %v", err)
	}

	// Using the Config value, create the DynamoDB,SNS client
	ddbClient = dynamodb.NewFromConfig(cfg)
	snsClient = sns.NewFromConfig(cfg)
}

// detail: https://docs.aws.amazon.com/zh_cn/lambda/latest/dg/with-kinesis.html
// input and output data: https://docs.aws.amazon.com/zh_cn/kinesisanalytics/latest/dev/how-it-works-output-lambda.html
// notice:
// Best effort, Kinesis Analytics Output is "at least once" delivery, meaning this lambda function can be invoked multiple times with the same item
// need Idempotent operation
// @TODO tracing https://docs.aws.amazon.com/zh_cn/lambda/latest/dg/golang-tracing.html
func Handler(ctx context.Context, kinesisAnalyticsEvent events.KinesisAnalyticsOutputDeliveryEvent) (responses events.KinesisAnalyticsOutputDeliveryResponse, err error) {
	responses = events.KinesisAnalyticsOutputDeliveryResponse{
		Records: make([]events.KinesisAnalyticsOutputDeliveryResponseRecord, len(kinesisAnalyticsEvent.Records)),
	}

	log.Printf("env TABLE_NAME:%s TOPIC_ARN:%s", eventDynamodbTable, eventSNSTopicArn)
	for i, record := range kinesisAnalyticsEvent.Records {
		responses.Records[i] = events.KinesisAnalyticsOutputDeliveryResponseRecord{
			RecordID: record.RecordID,
			Result:   events.KinesisAnalyticsOutputDeliveryOK,
		}

		dataBytes := record.Data

		log.Printf("%s Data = %s \n", record.RecordID, dataBytes)

		eventItem := &EventItem{}
		err = json.Unmarshal(dataBytes, eventItem)
		if err != nil {
			log.Printf("[WARNING] %s Data = %s can't decode by json error:%s \n", record.RecordID, dataBytes, err.Error())
			err = nil
			continue
		}
		tb := TableBasics{
			DynamoDbClient: ddbClient,
			TableName:      eventDynamodbTable,
		}
		err = tb.putItem(eventItem)
		if err != nil {
			log.Printf("[ERROR] %s Data = %s dynamoDb putItem: %v error:%s \n", record.RecordID, dataBytes, eventItem, err.Error())
			responses.Records[i].Result = events.KinesisAnalyticsOutputDeliveryFailed
			return responses, err
		}

		//go func() {
		input := &sns.PublishInput{
			Message:  aws.String(string(dataBytes)),
			TopicArn: aws.String(eventSNSTopicArn),
		}
		res, err := snsClient.Publish(context.TODO(), input)
		if err != nil {
			log.Printf("[WARNING] Data = %s can't send SNS err:%s \n", dataBytes, err.Error())
		} else {
			log.Printf("[INFO] Data = %s send SNS ok msgID:%s \n", dataBytes, *res.MessageId)
		}
		//}()
	}

	return responses, err
}

func main() {
	Init()
	lambda.Start(Handler)
}
