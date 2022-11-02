package main

import (
	"time"
	"user-behavior-analytics-cdk/infra"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awskinesis"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)
	kdsKdfS3, _ := KDSStack(app)

	WorkshopStack(app, kdsKdfS3.Stream())
	//WorkshopCICDPipelineStack(app)

	//RedshiftQuickSightStack(app)

	awscdk.Tags_Of(app).Add(jsii.String("version"), jsii.String("1.0"), nil)
	awscdk.Tags_Of(app).Add(jsii.String("project"), jsii.String("user-behavior-analytics"), nil)
	awscdk.Tags_Of(app).Add(jsii.String("role"), jsii.String("user behavior analytics streamimg and stroage"), nil)
	awscdk.Tags_Of(app).Add(jsii.String("synthTime"), jsii.String(time.Now().Format("2006-01-02 15:04:05.999")), nil)

	app.Synth(nil)
}

// dependecy kds
func WorkshopStack(app awscdk.App, eventStream awskinesis.Stream) {
	infra.NewCdkWsStack(app, "CDK-Workshop-Lambda-KDS-stack", &infra.CdkWsStackProps{
		StackProps: awscdk.StackProps{
			Env:         env(),
			StackName:   jsii.String("CDK-Workshop-lambda-KDS-stack"),
			Description: jsii.String("some cdk workshop demo constructs to test,then to use it"),
		},
		EventStream: eventStream,
	})
}

func WorkshopCICDPipelineStack(app awscdk.App) {
	infra.NewPipelineStack(app, "WorkshopCICDPipelineCdkStack", &infra.PipelineStackProps{
		StackProps: awscdk.StackProps{
			Env:         env(),
			StackName:   jsii.String("WorkshopCICDPipelineCdkStack"),
			Description: jsii.String("some cdk workshop pipleline demo"),
		},
	})
}

func RedshiftQuickSightStack(app awscdk.App) {
	infra.NewRedshiftQuicksightCdkStack(app, "RedshiftQuickSightStack", &infra.RedshiftQuicksightCdkStackProps{
		StackProps: awscdk.StackProps{
			Env:         env(),
			StackName:   jsii.String("RedshiftQuickSightStack"),
			Description: jsii.String("deploy Redshift and QuickSight"),
		},
	})
}

func KDSStack(app awscdk.App) (infra.KdsKdfS3Stack, awscdk.Stack) {
	kdsFirehoseS3Stack := infra.NewKdsKdfS3StackForUserBehaviorEvent(app, "KDS-KDF-S3-stack", &infra.KdsKdfS3StackProps{
		StackProps: awscdk.StackProps{
			Env:         env(),
			StackName:   jsii.String("KdsKdfS3StackForUserBehaviorEvent"),
			Description: jsii.String("aws kinesis data stream for firehose to s3"),
		},
	})

	stack := infra.NewKdsSqlKdaLambdaDynamoDBStack(app, "KDS-KDA-sql-Lambda-DynamoDB-stack", &infra.KdsSqlKdaLambdaDynamoDBStackProps{
		StackProps: awscdk.StackProps{
			Env:         env(),
			StackName:   jsii.String("KdsSqlKdaLambdaDynamoDBStackForUserBehaviorEvent"),
			Description: jsii.String("use aws kinesis data stream to analytics by sql"),
		},
		UseStream: kdsFirehoseS3Stack.Stream(),
	})

	return kdsFirehoseS3Stack, stack
}

// env determines the AWS environment (account+region) in which our stack is to
// be deployed. For more information see: https://docs.aws.amazon.com/cdk/latest/guide/environments.html
func env() *awscdk.Environment {
	// If unspecified, this stack will be "environment-agnostic".
	// Account/Region-dependent features and context lookups will not work, but a
	// single synthesized template can be deployed anywhere.
	//---------------------------------------------------------------------------
	return nil

	// Uncomment if you know exactly what account and region you want to deploy
	// the stack to. This is the recommendation for production stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String("123456789012"),
	//  Region:  jsii.String("us-east-1"),
	// }

	// Uncomment to specialize this stack for the AWS Account and Region that are
	// implied by the current CLI configuration. This is recommended for dev
	// stacks.
	//---------------------------------------------------------------------------
	// return &awscdk.Environment{
	//  Account: jsii.String(os.Getenv("CDK_DEFAULT_ACCOUNT")),
	//  Region:  jsii.String(os.Getenv("CDK_DEFAULT_REGION")),
	// }
}
