package main

import (
	"testing"
	"user-behavior-analytics-cdk/infra/lib"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/assertions"
	"github.com/aws/aws-cdk-go/awscdk/v2/awslambda"
	"github.com/aws/jsii-runtime-go"
	"github.com/google/go-cmp/cmp"
)

func TestHitCounterConstruct(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	stack := awscdk.NewStack(nil, nil, nil)

	// WHEN
	testFn := awslambda.NewFunction(stack, jsii.String("TestFunction"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("src/lambda/js-func/hello"), nil),
		Runtime: awslambda.Runtime_NODEJS_16_X(),
		Handler: jsii.String("hello.handler"),
	})
	lib.NewHitCounter(stack, "MyTestConstruct", &lib.HitCounterProps{
		Downstream: testFn,
	})

	// THEN
	template := assertions.Template_FromStack(stack, nil)
	template.ResourceCountIs(jsii.String("AWS::DynamoDB::Table"), jsii.Number(1))
}
func TestLambdaFunction(t *testing.T) {
	// GIVEN
	stack := awscdk.NewStack(nil, nil, nil)

	// WHEN
	testFn := awslambda.NewFunction(stack, jsii.String("TestFunction"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("src/lambda/js-func/hello"), nil),
		Runtime: awslambda.Runtime_NODEJS_16_X(),
		Handler: jsii.String("hello.handler"),
	})
	lib.NewHitCounter(stack, "MyTestConstruct", &lib.HitCounterProps{
		Downstream: testFn,
	})

	// THEN
	template := assertions.Template_FromStack(stack, nil)
	envCapture := assertions.NewCapture(nil)
	template.HasResourceProperties(jsii.String("AWS::Lambda::Function"), &map[string]any{
		"Environment": envCapture,
		"Handler":     "hitcounter.handler",
	})
	expectedEnv := &map[string]any{
		"Variables": map[string]any{
			"DOWNSTREAM_FUNCTION_NAME": map[string]any{
				"Ref": "TestFunction22AD90FC",
			},
			"HITS_TABLE_NAME": map[string]any{
				"Ref": "MyTestConstructHits24A357F0",
			},
		},
	}
	if !cmp.Equal(envCapture.AsObject(), expectedEnv) {
		t.Error(expectedEnv, envCapture.AsObject())
	}
}

func TestTableCreatedWithEncryption(t *testing.T) {
	defer jsii.Close()

	// GIVEN
	stack := awscdk.NewStack(nil, nil, nil)

	// WHEN
	testFn := awslambda.NewFunction(stack, jsii.String("TestFunction"), &awslambda.FunctionProps{
		Code:    awslambda.Code_FromAsset(jsii.String("src/lambda/js-func/hello"), nil),
		Runtime: awslambda.Runtime_NODEJS_16_X(),
		Handler: jsii.String("hello.handler"),
	})
	lib.NewHitCounter(stack, "MyTestConstruct", &lib.HitCounterProps{
		Downstream:   testFn,
		ReadCapacity: 10,
	})

	// THEN
	template := assertions.Template_FromStack(stack, nil)
	template.HasResourceProperties(jsii.String("AWS::DynamoDB::Table"), &map[string]any{
		"SSESpecification": map[string]any{
			"SSEEnabled": true,
		},
	})
}

func TestCantPassReadCapacity(t *testing.T) {
	defer jsii.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not throw ReadCapacity error")
		} else {
			t.Logf("%+v\n", r)
		}
	}()

	// GIVEN
	stack := awscdk.NewStack(nil, nil, nil)

	// THEN
	lib.NewHitCounter(stack, "MyTestConstruct", &lib.HitCounterProps{
		Downstream:   nil,
		ReadCapacity: 21,
	})
}

func TestCantPassStream(t *testing.T) {
	defer jsii.Close()
	defer func() {
		if r := recover(); r == nil {
			t.Error("Did not throw StreamObj error")
		} else {
			t.Logf("%+v\n", r)
		}
	}()

	// GIVEN
	stack := awscdk.NewStack(nil, nil, nil)

	// THEN
	lib.NewHitCounter(stack, "MyTestConstruct", &lib.HitCounterProps{
		Downstream:   nil,
		ReadCapacity: 7,
		EventStream:  nil,
	})
}
