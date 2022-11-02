package infra

import (
	"user-behavior-analytics-cdk/infra/lib"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskinesis"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type KdsKdfS3StackProps struct {
	awscdk.StackProps
}
type kdsKdfS3Stack struct {
	awscdk.Stack
	stream awskinesis.Stream
	bucket awss3.Bucket
}

func (m *kdsKdfS3Stack) Stream() awskinesis.Stream {
	return m.stream
}
func (m *kdsKdfS3Stack) Bucket() awss3.Bucket {
	return m.bucket
}

type KdsKdfS3Stack interface {
	awscdk.Stack
	Stream() awskinesis.Stream
	Bucket() awss3.Bucket
}

func NewKdsKdfS3StackForUserBehaviorEvent(scope constructs.Construct, id string, props *KdsKdfS3StackProps) KdsKdfS3Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}

	stack := awscdk.NewStack(scope, &id, &sprops)
	streamName := stack.Node().TryGetContext(jsii.String("kinesisDataStreamName")).(string)
	compressionFormat := stack.Node().TryGetContext(jsii.String("s3CompressionFormat")).(string)

	kdsFirehoseS3Construct := lib.NewKdsFirehoseS3Construct(stack, "KdsFirehoseS3Construct", &lib.KdsFirehoseS3Props{
		StreamName:        streamName,
		CompressionFormat: compressionFormat,
	})

	return &kdsKdfS3Stack{stream: kdsFirehoseS3Construct.Stream(), bucket: kdsFirehoseS3Construct.Bucket()}
}
