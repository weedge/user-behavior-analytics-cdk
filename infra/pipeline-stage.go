package infra

import (
	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/constructs-go/constructs/v10"
)

type WorkshopPipelineStageProps struct {
	awscdk.StageProps
}
type workshopPipelineStage struct {
	stage        awscdk.Stage
	hcGwUrl      awscdk.CfnOutput
	hcTvEndpoint awscdk.CfnOutput
}

type WorkshopPipelineStage interface {
	Stage() awscdk.Stage
	HcGwUrl() awscdk.CfnOutput
	HcTvEndpoint() awscdk.CfnOutput
}

func NewWorkshopPipelineStage(scope constructs.Construct, id string, props *WorkshopPipelineStageProps) WorkshopPipelineStage {
	var sprops awscdk.StageProps
	if props != nil {
		sprops = props.StageProps
	}
	stage := awscdk.NewStage(scope, &id, &sprops)

	workshopStack := NewCdkWsStack(stage, "WebService", nil)

	return &workshopPipelineStage{stage, workshopStack.HcGwUrl(), workshopStack.HcTvEndpoint()}
}

func (s *workshopPipelineStage) Stage() awscdk.Stage {
	return s.stage
}

func (s *workshopPipelineStage) HcGwUrl() awscdk.CfnOutput {
	return s.hcGwUrl
}

func (s *workshopPipelineStage) HcTvEndpoint() awscdk.CfnOutput {
	return s.hcTvEndpoint
}
