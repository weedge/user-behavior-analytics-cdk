package infra

import (
	"user-behavior-analytics-cdk/infra/lib"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapigateway"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdklabs/cdk-dynamo-table-viewer-go/dynamotableviewer"
)

type CdkWsStackProps struct {
	awscdk.StackProps
}
type cdkWsStack struct {
	awscdk.Stack
	hcGwUrl      awscdk.CfnOutput
	hcTvEndpoint awscdk.CfnOutput
}

type CdkWsStack interface {
	awscdk.Stack
	HcGwUrl() awscdk.CfnOutput
	HcTvEndpoint() awscdk.CfnOutput
}

func NewCdkWsStack(scope constructs.Construct, id string, props *CdkWsStackProps) CdkWsStack {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// The code that defines your stack goes here
	helloHandler := awslambda.NewFunction(stack, jsii.String("HelloHandler"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("src/lambda/js-func/hello"), nil),
		Runtime: awslambda.Runtime_NODEJS_16_X(),
		Handler: jsii.String("hello.handler"),
	})

	hitcounterObj := lib.NewHitCounter(stack, "HelloHitCounter", &lib.HitCounterProps{
		Downstream:   helloHandler,
		ReadCapacity: 7,
	})

	gateway := awsapigateway.NewLambdaRestApi(stack, jsii.String("Endpoint"), &awsapigateway.LambdaRestApiProps{
		Handler: hitcounterObj.Handler(),
	})

	tv := dynamotableviewer.NewTableViewer(stack, jsii.String("ViewHitCounter"), &dynamotableviewer.TableViewerProps{
		Title: jsii.String("Hello Hits"),
		Table: hitcounterObj.Table(),
	})

	hcGwUrl := awscdk.NewCfnOutput(stack, jsii.String("GatewayUrl"), &awscdk.CfnOutputProps{
		Value: gateway.Url(),
	})

	hcTvEndpoint := awscdk.NewCfnOutput(stack, jsii.String("TableViewerUrl"), &awscdk.CfnOutputProps{
		Value: tv.Endpoint(),
	})

	return &cdkWsStack{stack, hcGwUrl, hcTvEndpoint}
}

func (s *cdkWsStack) HcGwUrl() awscdk.CfnOutput {
	return s.hcGwUrl
}

func (s *cdkWsStack) HcTvEndpoint() awscdk.CfnOutput {
	return s.hcTvEndpoint
}
