package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/aws/aws-sdk-go-v2/service/autoscaling/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedAutoscalingService(mock *mockAutoscalingClient) aws.NoclickopsAutoscalingService {
	return aws.NoclickopsAutoscalingService{
		Clients: []aws.NoclickopsAutoscalingClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "autoscaling"},
	}
}

func TestGetAllAutoScalingGroups_InstancesNotRequested(t *testing.T) {
	mock := &mockAutoscalingClient{
		describeAutoScalingGroupsFn: func(_ context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, _ ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
			if params.IncludeInstances == nil || *params.IncludeInstances {
				t.Error("expected IncludeInstances to be false")
			}
			return &autoscaling.DescribeAutoScalingGroupsOutput{}, nil
		},
	}
	client := getMockedAutoscalingService(mock)
	client.GetAllAutoScalingGroups()
}

func TestGetAllAutoScalingGroups_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockAutoscalingClient{
		describeAutoScalingGroupsFn: func(_ context.Context, params *autoscaling.DescribeAutoScalingGroupsInput, _ ...func(*autoscaling.Options)) (*autoscaling.DescribeAutoScalingGroupsOutput, error) {
			callCount++
			if callCount == 1 {
				return &autoscaling.DescribeAutoScalingGroupsOutput{
					NextToken:         ptr("next"),
					AutoScalingGroups: []types.AutoScalingGroup{{AutoScalingGroupName: ptr("asg-1")}},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken, expected 'next' got '%v'", params.NextToken)
			}
			return &autoscaling.DescribeAutoScalingGroupsOutput{
				AutoScalingGroups: []types.AutoScalingGroup{{AutoScalingGroupName: ptr("asg-2")}},
			}, nil
		},
	}
	client := getMockedAutoscalingService(mock)
	got := client.GetAllAutoScalingGroups()
	expected := []common.Resource{
		{TerraformID: "asg-1", ResourceType: common.Autoscaling_group, Region: "eu-west-1"},
		{TerraformID: "asg-2", ResourceType: common.Autoscaling_group, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeAutoScalingGroups, got %d", callCount)
	}
}
