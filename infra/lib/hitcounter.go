package lib

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskinesis"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type HitCounterProps struct {
	Downstream   awslambda.IFunction
	ReadCapacity float64
	EventStream  awskinesis.Stream
}

type hitCounter struct {
	constructs.Construct
	handler awslambda.IFunction
	table   awsdynamodb.Table
}

type HitCounter interface {
	constructs.Construct
	Handler() awslambda.IFunction
	Table() awsdynamodb.Table
}

func NewHitCounter(scope constructs.Construct, id string, props *HitCounterProps) HitCounter {
	if props.ReadCapacity < 5 || props.ReadCapacity > 20 {
		panic("ReadCapacity must be between 5 and 20")
	}
	if props.EventStream == nil || props.EventStream.StreamName() == nil {
		panic("eventStream is nil")
	}

	this := constructs.NewConstruct(scope, &id)

	table := awsdynamodb.NewTable(this, jsii.String("Hits"), &awsdynamodb.TableProps{
		PartitionKey:  &awsdynamodb.Attribute{Name: jsii.String("path"), Type: awsdynamodb.AttributeType_STRING},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		Encryption:    awsdynamodb.TableEncryption_AWS_MANAGED,
		ReadCapacity:  &props.ReadCapacity,
	})

	handler := awslambda.NewFunction(this, jsii.String("HitCounterHandler"), &awslambda.FunctionProps{
		Runtime: awslambda.Runtime_NODEJS_16_X(),
		Handler: jsii.String("hitcounter.handler"),
		Code:    awslambda.Code_FromAsset(jsii.String("src/lambda/js-func/hitcounter"), nil),
		Environment: &map[string]*string{
			"DOWNSTREAM_FUNCTION_NAME": props.Downstream.FunctionName(),
			"HITS_TABLE_NAME":          table.TableName(),
			"HITS_STREAM_NAME":         props.EventStream.StreamName(),
		},
	})

	table.GrantReadWriteData(handler)
	props.Downstream.GrantInvoke(handler)
	props.EventStream.GrantWrite(handler)

	return &hitCounter{this, handler, table}
}

func (h *hitCounter) Handler() awslambda.IFunction {
	return h.handler
}

func (h *hitCounter) Table() awsdynamodb.Table {
	return h.table
}
