package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedIAMService(mock *mockIAMClient) aws.NoclickopsIAMService {
	return aws.NoclickopsIAMService{
		Clients: []aws.NoclickopsIAMClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "global"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: true, ServiceName: "ec2"},
	}
}

func TestListPolicies_PaginationFollowed(t *testing.T) {
	callCount := 0

	mock := &mockIAMClient{
		listPoliciesFn: func(_ context.Context, params *iam.ListPoliciesInput, _ ...func(*iam.Options)) (*iam.ListPoliciesOutput, error) {
			callCount++
			if callCount == 1 {
				return &iam.ListPoliciesOutput{
					IsTruncated: true,
					Marker:      ptr("next"),
					Policies: []types.Policy{
						{Arn: ptr("arn:policy_1")},
					},
				}, nil
			}
			// assert pagination params were passed through correctly
			if params.Marker == nil || *params.Marker != "next" {
				return nil, fmt.Errorf("Wrong Marker, expected 'next' got '%v'", *params.Marker)
			}
			return &iam.ListPoliciesOutput{
				IsTruncated: false,
				Policies: []types.Policy{
					{Arn: ptr("arn:policy_2")},
				},
			}, nil
		},
	}
	client := getMockedIAMService(mock)
	ids := client.GetAllPoliciesArns()
	expected := []common.Resource{
		{Arn: "arn:policy_1", TerraformID: "arn:policy_1", ResourceType: common.IAM_policy, Region: "global"},
		{Arn: "arn:policy_2", TerraformID: "arn:policy_2", ResourceType: common.IAM_policy, Region: "global"},
	}
	if diff := cmp.Diff(ids, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, ids)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListPolicies, got %d", callCount)
	}
}

func TestGetAllIAMUsers_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockIAMClient{
		listUsersFn: func(_ context.Context, params *iam.ListUsersInput, _ ...func(*iam.Options)) (*iam.ListUsersOutput, error) {
			callCount++
			if callCount == 1 {
				return &iam.ListUsersOutput{
					IsTruncated: true,
					Marker:      ptr("next"),
					Users: []types.User{
						{UserName: ptr("user_1"), Arn: ptr("arn:user_1")},
					},
				}, nil
			}
			if params.Marker == nil || *params.Marker != "next" {
				return nil, fmt.Errorf("wrong Marker, expected 'next' got '%v'", params.Marker)
			}
			return &iam.ListUsersOutput{
				IsTruncated: false,
				Users: []types.User{
					{UserName: ptr("user_2"), Arn: ptr("arn:user_2")},
				},
			}, nil
		},
	}
	client := getMockedIAMService(mock)
	got := client.GetAllIAMUsers()
	expected := []common.Resource{
		{Arn: "arn:user_1", TerraformID: "user_1", ResourceType: common.IAM_user, Region: "global"},
		{Arn: "arn:user_2", TerraformID: "user_2", ResourceType: common.IAM_user, Region: "global"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListUsers, got %d", callCount)
	}
}

func TestGetAllIAMGroups_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockIAMClient{
		listGroupsFn: func(_ context.Context, params *iam.ListGroupsInput, _ ...func(*iam.Options)) (*iam.ListGroupsOutput, error) {
			callCount++
			if callCount == 1 {
				return &iam.ListGroupsOutput{
					IsTruncated: true,
					Marker:      ptr("next"),
					Groups: []types.Group{
						{GroupName: ptr("group_1"), Arn: ptr("arn:group_1")},
					},
				}, nil
			}
			if params.Marker == nil || *params.Marker != "next" {
				return nil, fmt.Errorf("wrong Marker, expected 'next' got '%v'", params.Marker)
			}
			return &iam.ListGroupsOutput{
				IsTruncated: false,
				Groups: []types.Group{
					{GroupName: ptr("group_2"), Arn: ptr("arn:group_2")},
				},
			}, nil
		},
	}
	client := getMockedIAMService(mock)
	got := client.GetAllIAMGroups()
	expected := []common.Resource{
		{Arn: "arn:group_1", TerraformID: "group_1", ResourceType: common.IAM_group, Region: "global"},
		{Arn: "arn:group_2", TerraformID: "group_2", ResourceType: common.IAM_group, Region: "global"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListGroups, got %d", callCount)
	}
}
