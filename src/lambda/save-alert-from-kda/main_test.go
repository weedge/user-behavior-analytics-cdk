package main

import (
	"context"
	"os"
	"reflect"
	"testing"

	"github.com/aws/aws-lambda-go/events"
)

// Tips: env params need to set(TOPIC_ARN, TABLE_NAME) before run test if need env params
func TestMain(m *testing.M) {
	os.Setenv("TABLE_NAME", "test")
	os.Setenv("TOPIC_ARN", "test")
	Init()
}

func TestHandler(t *testing.T) {
	type args struct {
		ctx                   context.Context
		kinesisAnalyticsEvent events.KinesisAnalyticsOutputDeliveryEvent
	}
	tests := []struct {
		name          string
		args          args
		wantResponses events.KinesisAnalyticsOutputDeliveryResponse
		wantErr       bool
	}{
		// TODO: Add test data from lambda test case.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotResponses, err := Handler(tt.args.ctx, tt.args.kinesisAnalyticsEvent)
			if (err != nil) != tt.wantErr {
				t.Errorf("Handler() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(gotResponses, tt.wantResponses) {
				t.Errorf("Handler() = %v, want %v", gotResponses, tt.wantResponses)
			}
		})
	}
}
