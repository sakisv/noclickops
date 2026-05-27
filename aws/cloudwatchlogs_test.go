package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedCloudwatchLogsService(mock *mockCloudwatchLogsClient) aws.NoclickopsCloudwatchLogsService {
	return aws.NoclickopsCloudwatchLogsService{
		Clients: []aws.NoclickopsCloudwatchLogsClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "cloudwatchlogs"},
	}
}

func TestGetLogGroups_BasicCase(t *testing.T) {
	mock := &mockCloudwatchLogsClient{
		listLogGroupsFn: func(_ context.Context, _ *cloudwatchlogs.ListLogGroupsInput, _ ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListLogGroupsOutput, error) {
			return &cloudwatchlogs.ListLogGroupsOutput{
				LogGroups: []types.LogGroupSummary{
					{LogGroupArn: ptr("arn:aws:logs:eu-west-1:123456789012:log-group:group-1:*"), LogGroupName: ptr("group-1")},
					{LogGroupArn: ptr("arn:aws:logs:eu-west-1:123456789012:log-group:group-2:*"), LogGroupName: ptr("group-2")},
				},
			}, nil
		},
	}
	svc := getMockedCloudwatchLogsService(mock)
	got := svc.GetLogGroups()
	expected := []common.Resource{
		{Arn: "arn:aws:logs:eu-west-1:123456789012:log-group:group-1:*", TerraformID: "group-1", ResourceType: common.Cloudwatchlogs_log_group, Region: "eu-west-1"},
		{Arn: "arn:aws:logs:eu-west-1:123456789012:log-group:group-2:*", TerraformID: "group-2", ResourceType: common.Cloudwatchlogs_log_group, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetLogGroups_NoLogGroups(t *testing.T) {
	mock := &mockCloudwatchLogsClient{
		listLogGroupsFn: func(_ context.Context, _ *cloudwatchlogs.ListLogGroupsInput, _ ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListLogGroupsOutput, error) {
			return &cloudwatchlogs.ListLogGroupsOutput{}, nil
		},
	}
	svc := getMockedCloudwatchLogsService(mock)
	got := svc.GetLogGroups()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetLogGroups_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockCloudwatchLogsClient{
		listLogGroupsFn: func(_ context.Context, params *cloudwatchlogs.ListLogGroupsInput, _ ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListLogGroupsOutput, error) {
			callCount++
			if callCount == 1 {
				return &cloudwatchlogs.ListLogGroupsOutput{
					LogGroups: []types.LogGroupSummary{
						{LogGroupArn: ptr("arn:aws:logs:eu-west-1:123456789012:log-group:group-1:*"), LogGroupName: ptr("group-1")},
					},
					NextToken: ptr("next-page-token"),
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next-page-token" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next-page-token', got %v", params.NextToken)
			}
			return &cloudwatchlogs.ListLogGroupsOutput{
				LogGroups: []types.LogGroupSummary{
					{LogGroupArn: ptr("arn:aws:logs:eu-west-1:123456789012:log-group:group-2:*"), LogGroupName: ptr("group-2")},
				},
			}, nil
		},
	}
	svc := getMockedCloudwatchLogsService(mock)
	got := svc.GetLogGroups()
	expected := []common.Resource{
		{Arn: "arn:aws:logs:eu-west-1:123456789012:log-group:group-1:*", TerraformID: "group-1", ResourceType: common.Cloudwatchlogs_log_group, Region: "eu-west-1"},
		{Arn: "arn:aws:logs:eu-west-1:123456789012:log-group:group-2:*", TerraformID: "group-2", ResourceType: common.Cloudwatchlogs_log_group, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListLogGroups, got %d", callCount)
	}
}

func TestGetLogGroups_ErrorIsSkipped(t *testing.T) {
	mock := &mockCloudwatchLogsClient{
		listLogGroupsFn: func(_ context.Context, _ *cloudwatchlogs.ListLogGroupsInput, _ ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListLogGroupsOutput, error) {
			return nil, fmt.Errorf("access denied")
		},
	}
	svc := getMockedCloudwatchLogsService(mock)
	got := svc.GetLogGroups()
	if len(got) != 0 {
		t.Errorf("expected empty result on error, got %v", got)
	}
}
