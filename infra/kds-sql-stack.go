package infra

import (
	"os"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsdynamodb"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsiam"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskinesis"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskinesisanalytics"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskinesisfirehose"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssnssubscriptions"

	awscdklambdago "github.com/aws/aws-cdk-go/awscdklambdagoalpha/v2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdklabs/cdk-dynamo-table-viewer-go/dynamotableviewer"
)

type UserBehaviorAnalyticsKDSCdkStackProps struct {
	awscdk.StackProps
}

func NewUserBehaviorAnalyticsKDSCdkStack(scope constructs.Construct, id string, props *UserBehaviorAnalyticsKDSCdkStackProps) awscdk.Stack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// eventStream is a raw kinesis data stream
	eventStream := awskinesis.NewStream(stack, jsii.String("EventStream"), nil)

	// outPut the stream name so can connect our script to this stream
	awscdk.NewCfnOutput(stack, jsii.String("EventStreamName"), &awscdk.CfnOutputProps{
		Value: eventStream.StreamName(),
	})

	// ==== history/cold mudule ====
	// S3 bucket that serve as the desc for raw compressed data in ODS
	rawDataBucket := awss3.NewBucket(stack, jsii.String("RawDataBucket"), &awss3.BucketProps{
		RemovalPolicy:     awscdk.RemovalPolicy_DESTROY, // REMOVE FOR PRODUCTION
		AutoDeleteObjects: jsii.Bool(true),              // REMOVE FOR PROUCTION
	})

	firehoseRole := awsiam.NewRole(stack, jsii.String("firehoseRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("firehose.amazonaws.com"), nil),
	})

	eventStream.GrantRead(firehoseRole)
	eventStream.Grant(firehoseRole, jsii.String("kinesis:DescribeStream"))
	rawDataBucket.GrantWrite(firehoseRole, nil)

	firehoseDeliveryStreamToS3 := awskinesisfirehose.NewCfnDeliveryStream(stack, jsii.String("FirehoseDeliveryStreamToS3"), &awskinesisfirehose.CfnDeliveryStreamProps{
		DeliveryStreamName: jsii.String("RawDataStreamToS3"),
		DeliveryStreamType: jsii.String("KinesisStreamAsSource"),
		KinesisStreamSourceConfiguration: &awskinesisfirehose.CfnDeliveryStream_KinesisStreamSourceConfigurationProperty{
			KinesisStreamArn: eventStream.StreamArn(),
			RoleArn:          firehoseRole.RoleArn(),
		},
		S3DestinationConfiguration: &awskinesisfirehose.CfnDeliveryStream_S3DestinationConfigurationProperty{
			BucketArn: rawDataBucket.BucketArn(),
			RoleArn:   firehoseRole.RoleArn(),
			BufferingHints: &awskinesisfirehose.CfnDeliveryStream_BufferingHintsProperty{
				IntervalInSeconds: jsii.Number(60),
				SizeInMBs:         jsii.Number(64),
			},
			CompressionFormat: jsii.String("GZIP"),
			//CompressionFormat: jsii.String("UNCOMPRESSED"),
			EncryptionConfiguration: &awskinesisfirehose.CfnDeliveryStream_EncryptionConfigurationProperty{
				NoEncryptionConfig: jsii.String("NoEncryption"),
			},
			Prefix: jsii.String("raw/"),
		},
	})

	// Ensures firehose role is created before create a Kinesis Firehose
	firehoseDeliveryStreamToS3.Node().AddDependency(firehoseRole)

	// ==== hot mudule ====

	// The DynamoDB table that stores user behavior abnormal event result by kinesis analytic app through lambda function to write
	userBeHaviorAbnormalTable := awsdynamodb.NewTable(stack, jsii.String("UserBehaviorAbnormalEventTable"), &awsdynamodb.TableProps{
		PartitionKey: &awsdynamodb.Attribute{
			Name: jsii.String("eventId"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		SortKey: &awsdynamodb.Attribute{
			Name: jsii.String("createdAt"),
			Type: awsdynamodb.AttributeType_STRING,
		},
		RemovalPolicy: awscdk.RemovalPolicy_DESTROY,
		TableName:     jsii.String("UserBeHaviorAbnormalEvent"), //biz define table name
	})

	//table viewer is demo construct of a web app for dynamodb table display, use aws api gateway and serverless lambda function
	dynamotableviewer.NewTableViewer(stack, jsii.String("UserBehaviorAbnormalView"), &dynamotableviewer.TableViewerProps{
		Table:  userBeHaviorAbnormalTable,
		SortBy: jsii.String("-createdAt"),
		Title:  jsii.String("User Behavior Abnormal View"),
	})

	abnormalEventNoticationTopic := awssns.NewTopic(stack, jsii.String("AbnormalEventNotication"), &awssns.TopicProps{
		DisplayName: jsii.String("AbnormalEventAlertNotication"),
	})

	// new email subscription to alert
	// u can new lambda subscription to send feishu or dingTalk alert
	abnormalEventNoticationTopic.AddSubscription(awssnssubscriptions.NewEmailSubscription(
		jsii.String("weege007@gmail.com"), // biz define alert email
		nil,
	))

	// Lambda function that reads output from our kinesis analytic app and save to DynamoDB table
	// and alert abnormal events
	saveAlertLambda := awscdklambdago.NewGoFunction(stack, jsii.String("UserBehaviorAnalytics-SaveAlertFunc"), &awscdklambdago.GoFunctionProps{
		FunctionName: jsii.String("UserBehaviorAnalytics-SaveAlertFunc"),
		Description:  jsii.String("reads output from our kinesis analytic app and save to DynamoDB table and write to sns for email alert"),
		Entry:        jsii.String("src/lambda/save-alert-from-kad"),
		Environment: &map[string]*string{
			"TABLE_NAME": userBeHaviorAbnormalTable.TableName(),
			"TOPIC_ARN":  abnormalEventNoticationTopic.TopicArn(),
		},
	})

	/*
		saveAlertLambda := awslambda.NewFunction(stack, jsii.String("LambdaSaveAlertFunction"), &awslambda.FunctionProps{
			Runtime: awslambda.Runtime_GO_1_X(),
			Code:    awslambda.Code_FromAsset(jsii.String("src/lambda/save-alert-from-kad"), &awss3assets.AssetOptions{}),
			Handler: jsii.String("lambdaHandler"), // need go build -ldflags="-s -w" -o lambdaHandler in lambda func dir with GOOS GOARCH params
			Environment: &map[string]*string{
				"TABLE_NAME": userBeHaviorAbnormalTable.TableName(),
				"TOPIC_ARN":  abnormalEventNoticationTopic.TopicArn(),
			},
		})
	*/
	abnormalEventNoticationTopic.GrantPublish(saveAlertLambda)
	userBeHaviorAbnormalTable.GrantReadWriteData(saveAlertLambda)

	// create stream analytics role for kinesis analytics app
	streamToAnalyticsRole := awsiam.NewRole(stack, jsii.String("streamToAnalyticsRole"), &awsiam.RoleProps{
		AssumedBy: awsiam.NewServicePrincipal(jsii.String("kinesisanalytics.amazonaws.com"), nil),
	})
	streamToAnalyticsRole.AddToPolicy(awsiam.NewPolicyStatement(&awsiam.PolicyStatementProps{
		Actions:   &[]*string{jsii.String("kinesis:*"), jsii.String("lambda:*")},
		Resources: &[]*string{eventStream.StreamArn(), saveAlertLambda.FunctionArn()},
	}))

	// create kinesis analytics app use sql(old version) from kinesis data stream
	// for abnormality event alert
	dir, _ := os.Getwd()
	sqlCode, err := os.ReadFile(dir + "/src/kinesis-analytics-sql/filter-abnormality-event.sql")
	if err != nil {
		panic(err.Error())
	}
	kinesisAnalyticsAppForAbnormalityEvent := awskinesisanalytics.NewCfnApplication(stack, jsii.String("KinesisAnalyticsApplication"), &awskinesisanalytics.CfnApplicationProps{
		ApplicationName:        jsii.String("abnormality-event-detector"),
		ApplicationDescription: jsii.String("use kinesis sql to analytics filter abnormality event"),
		ApplicationCode:        jsii.String(string(sqlCode)),
		// https://docs.aws.amazon.com/zh_cn/kinesisanalytics/latest/dev/how-it-works-input.html
		Inputs: []awskinesisanalytics.CfnApplication_InputProperty{
			{
				NamePrefix: jsii.String("SOURCE_SQL_STREAM"),
				KinesisStreamsInput: &awskinesisanalytics.CfnApplicationOutput_KinesisStreamsOutputProperty{
					ResourceArn: eventStream.StreamArn(),
					RoleArn:     streamToAnalyticsRole.RoleArn(),
				},
				InputParallelism: &awskinesisanalytics.CfnApplication_InputParallelismProperty{
					Count: jsii.Number(1),
				},
				// https://docs.aws.amazon.com/zh_cn/kinesisanalytics/latest/dev/about-json-path.html
				InputSchema: &awskinesisanalytics.CfnApplication_InputSchemaProperty{
					RecordFormat: &awskinesisanalytics.CfnApplication_RecordFormatProperty{
						RecordFormatType: jsii.String("JSON"),
						MappingParameters: &awskinesisanalytics.CfnApplication_MappingParametersProperty{
							JsonMappingParameters: &awskinesisanalytics.CfnApplication_JSONMappingParametersProperty{
								RecordRowPath: jsii.String("$"),
							},
						},
					},
					RecordEncoding: jsii.String("UTF-8"),
					// https://docs.aws.amazon.com/zh_cn/kinesisanalytics/latest/dev/sch-mapping.html
					RecordColumns: []awskinesisanalytics.CfnApplication_RecordColumnProperty{
						{
							Name:    jsii.String("eventId"),
							SqlType: jsii.String("VARCHAR(64)"),
							Mapping: jsii.String("$.eventId"),
						},
						{
							Name:    jsii.String("action"),
							SqlType: jsii.String("VARCHAR(256)"),
							Mapping: jsii.String("$.action"),
						},
						{
							Name:    jsii.String("userId"),
							SqlType: jsii.String("VARCHAR(64)"),
							Mapping: jsii.String("$.userId"),
						},
						{
							Name:    jsii.String("objectId"),
							SqlType: jsii.String("VARCHAR(64)"),
							Mapping: jsii.String("$.objectId"),
						},
						{
							Name:    jsii.String("bizId"),
							SqlType: jsii.String("VARCHAR(64)"),
							Mapping: jsii.String("$.bizId"),
						},
						{
							Name:    jsii.String("errorMsg"),
							SqlType: jsii.String("VARCHAR(1024)"),
							Mapping: jsii.String("$.errorMsg"),
						},
						{
							Name:    jsii.String("createdAt"),
							SqlType: jsii.String("VARCHAR(32)"),
							Mapping: jsii.String("$.createdAt"),
						},
					},
				},
			},
		},
	})
	kinesisAnalyticsAppForAbnormalityEvent.Node().AddDependency(streamToAnalyticsRole)
	kinesisAnalyticsAppOutput := awskinesisanalytics.NewCfnApplicationOutput(stack, jsii.String("KinesisAnalyticsApplicationOutPut"), &awskinesisanalytics.CfnApplicationOutputProps{
		ApplicationName: jsii.String("abnormality-event-detector"),
		Output: &awskinesisanalytics.CfnApplicationOutput_OutputProperty{
			Name: jsii.String("DESTINATION_SQL_STREAM"),
			DestinationSchema: &awskinesisanalytics.CfnApplicationOutput_DestinationSchemaProperty{
				RecordFormatType: jsii.String("JSON"),
			},
			LambdaOutput: &awskinesisanalytics.CfnApplicationOutput_LambdaOutputProperty{
				ResourceArn: saveAlertLambda.FunctionArn(),
				RoleArn:     streamToAnalyticsRole.RoleArn(),
			},
			//KinesisFirehoseOutput: nil,
			//KinesisStreamsOutput:  nil,
		},
	})
	kinesisAnalyticsAppOutput.Node().AddDependency(kinesisAnalyticsAppForAbnormalityEvent)

	return stack
}
