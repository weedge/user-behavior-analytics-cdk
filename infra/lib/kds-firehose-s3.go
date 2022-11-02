package lib

import (
	"strings"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskinesis"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskinesisfirehose"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

type KdsFirehoseS3Props struct {
	StreamName        string
	CompressionFormat string
}

type kdsFirehoseS3Construct struct {
	constructs.Construct
	stream awskinesis.Stream
	bucket awss3.Bucket
}

func (m *kdsFirehoseS3Construct) Stream() awskinesis.Stream {
	return m.stream
}
func (m *kdsFirehoseS3Construct) Bucket() awss3.Bucket {
	return m.bucket
}

type IKdsFirehoseS3Construct interface {
	constructs.Construct
	Stream() awskinesis.Stream
	Bucket() awss3.Bucket
}

func NewKdsFirehoseS3Construct(scope constructs.Construct, id string, props *KdsFirehoseS3Props) IKdsFirehoseS3Construct {
	if len(strings.Trim(props.StreamName, " ")) == 0 {
		panic("StreamName is empty")
	}

	this := constructs.NewConstruct(scope, &id)

	// new kinesis data stream
	dataStream := awskinesis.NewStream(this, jsii.String(props.StreamName), nil)

	// outPut the stream name so can connect our script to this stream
	awscdk.NewCfnOutput(this, jsii.String("dataStreamName"), &awscdk.CfnOutputProps{
		Value: dataStream.StreamName(),
	})

	// S3 bucket that serve as the desc
	rawDataBucket := awss3.NewBucket(this, jsii.String("RawDataBucket"), &awss3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY, // REMOVE FOR PRODUCTION
		AutoDeleteObjects: jsii.Bool(true),              // REMOVE FOR PROUCTION
	})

	firehoseRole := awsiam.NewRole(this, jsii.String("firehoseRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("firehose.amazonaws.com"), nil),
	})

	dataStream.GrantRead(firehoseRole)
	dataStream.Grant(firehoseRole, jsii.String("kinesis:DescribeStream"))
	rawDataBucket.GrantWrite(firehoseRole, nil)

	firehoseDeliveryStreamToS3 := awskinesisfirehose.NewCfnDeliveryStream(this, jsii.String("FirehoseDeliveryStreamToS3"), &awskinesisfirehose.CfnDeliveryStreamProps{
		//DeliveryStreamName: jsii.String("RawDataStreamToS3"),
		DeliveryStreamType: jsii.String("KinesisStreamAsSource"),
		KinesisStreamSourceConfiguration: &awskinesisfirehose.CfnDeliveryStream_KinesisStreamSourceConfigurationProperty{
			KinesisStreamArn: dataStream.StreamArn(),
			RoleArn:          firehoseRole.RoleArn(),
		},
		S3DestinationConfiguration: &awskinesisfirehose.CfnDeliveryStream_S3DestinationConfigurationProperty{
			BucketArn: rawDataBucket.BucketArn(),
			RoleArn:   firehoseRole.RoleArn(),
			BufferingHints: &awskinesisfirehose.CfnDeliveryStream_BufferingHintsProperty{
				IntervalInSeconds: jsii.Number(60),
				SizeInMBs:         jsii.Number(64),
			},
			CompressionFormat: jsii.String(props.CompressionFormat),
			//CompressionFormat: jsii.String("UNCOMPRESSED"),
			EncryptionConfiguration: &awskinesisfirehose.CfnDeliveryStream_EncryptionConfigurationProperty{
				NoEncryptionConfig: jsii.String("NoEncryption"),
			},
			Prefix: jsii.String("raw/"),
		},
	})

	// Ensures firehose role is created before create a Kinesis Firehose
	firehoseDeliveryStreamToS3.Node().AddDependency(firehoseRole)

	return &kdsFirehoseS3Construct{stream: dataStream, bucket: rawDataBucket}
}
