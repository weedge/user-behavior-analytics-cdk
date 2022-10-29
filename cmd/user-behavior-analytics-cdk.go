package main

import (
	"time"
	"user-behavior-analytics-cdk/infra"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	KDSStack(app)

	app.Synth(nil)
}

func KDSStack(app awscdk.App) {
	stack := infra.NewUserBehaviorAnalyticsKDSCdkStack(app, "UserBehaviorAnalyticsCdkStack", &infra.UserBehaviorAnalyticsKDSCdkStackProps{
		StackProps: awscdk.StackProps{
			Env:         env(),
			StackName:   jsii.String("UserBehaviorAnalyticsCdkStackKinesisDataStream"),
			Description: jsii.String("use aws kinesis data stream to analytics"),
		},
	})

	awscdk.Tags_Of(stack).Add(jsii.String("version"), jsii.String("1.0"), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("project"), jsii.String("user-behavior-analytics"), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("role"), jsii.String("user behavior analytics streamimg and stroage"), nil)
	awscdk.Tags_Of(stack).Add(jsii.String("synthTime"), jsii.String(time.Now().Format("2006-01-02 15:04:05.999")), nil)
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
