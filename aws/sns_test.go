package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/aws/aws-sdk-go-v2/service/sns/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedSNSService(mock *mockSNSClient) aws.NoclickopsSNSService {
	return aws.NoclickopsSNSService{
		Clients: []aws.NoclickopsSNSClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "sns"},
	}
}

func TestGetAllTopics_BasicCase(t *testing.T) {
	mock := &mockSNSClient{
		listTopicsFn: func(_ context.Context, _ *sns.ListTopicsInput, _ ...func(*sns.Options)) (*sns.ListTopicsOutput, error) {
			return &sns.ListTopicsOutput{
				Topics: []types.Topic{
					{TopicArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-1")},
					{TopicArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-2")},
				},
			}, nil
		},
	}
	svc := getMockedSNSService(mock)
	got := svc.GetAllTopics()
	expected := []common.Resource{
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-1", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-1", ResourceType: common.SNS_topic, Region: "eu-west-1"},
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-2", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-2", ResourceType: common.SNS_topic, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetAllTopics_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockSNSClient{
		listTopicsFn: func(_ context.Context, params *sns.ListTopicsInput, _ ...func(*sns.Options)) (*sns.ListTopicsOutput, error) {
			callCount++
			if callCount == 1 {
				return &sns.ListTopicsOutput{
					Topics:    []types.Topic{{TopicArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-1")}},
					NextToken: ptr("next-topic"),
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next-topic" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next-topic', got %v", params.NextToken)
			}
			return &sns.ListTopicsOutput{
				Topics: []types.Topic{{TopicArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-2")}},
			}, nil
		},
	}
	svc := getMockedSNSService(mock)
	got := svc.GetAllTopics()
	expected := []common.Resource{
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-1", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-1", ResourceType: common.SNS_topic, Region: "eu-west-1"},
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-2", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-2", ResourceType: common.SNS_topic, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListTopics, got %d", callCount)
	}
}

func TestGetAllTopics_NoTopics(t *testing.T) {
	mock := &mockSNSClient{
		listTopicsFn: func(_ context.Context, _ *sns.ListTopicsInput, _ ...func(*sns.Options)) (*sns.ListTopicsOutput, error) {
			return &sns.ListTopicsOutput{}, nil
		},
	}
	svc := getMockedSNSService(mock)
	got := svc.GetAllTopics()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetAllSubscriptions_BasicCase(t *testing.T) {
	mock := &mockSNSClient{
		listSubscriptionsFn: func(_ context.Context, _ *sns.ListSubscriptionsInput, _ ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error) {
			return &sns.ListSubscriptionsOutput{
				Subscriptions: []types.Subscription{
					{SubscriptionArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa")},
					{SubscriptionArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-1:sub-bbb")},
				},
			}, nil
		},
	}
	svc := getMockedSNSService(mock)
	got := svc.GetAllSubscriptions()
	expected := []common.Resource{
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa", ResourceType: common.SNS_subscription, Region: "eu-west-1"},
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-bbb", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-bbb", ResourceType: common.SNS_subscription, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetAllSubscriptions_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockSNSClient{
		listSubscriptionsFn: func(_ context.Context, params *sns.ListSubscriptionsInput, _ ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error) {
			callCount++
			if callCount == 1 {
				return &sns.ListSubscriptionsOutput{
					Subscriptions: []types.Subscription{{SubscriptionArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa")}},
					NextToken:     ptr("next-sub"),
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next-sub" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next-sub', got %v", params.NextToken)
			}
			return &sns.ListSubscriptionsOutput{
				Subscriptions: []types.Subscription{{SubscriptionArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-1:sub-bbb")}},
			}, nil
		},
	}
	svc := getMockedSNSService(mock)
	got := svc.GetAllSubscriptions()
	expected := []common.Resource{
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa", ResourceType: common.SNS_subscription, Region: "eu-west-1"},
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-bbb", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-bbb", ResourceType: common.SNS_subscription, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListSubscriptions, got %d", callCount)
	}
}

func TestGetAllSubscriptions_NoSubscriptions(t *testing.T) {
	mock := &mockSNSClient{
		listSubscriptionsFn: func(_ context.Context, _ *sns.ListSubscriptionsInput, _ ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error) {
			return &sns.ListSubscriptionsOutput{}, nil
		},
	}
	svc := getMockedSNSService(mock)
	got := svc.GetAllSubscriptions()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetAllResources_CombinesTopicsAndSubscriptions(t *testing.T) {
	mock := &mockSNSClient{
		listTopicsFn: func(_ context.Context, _ *sns.ListTopicsInput, _ ...func(*sns.Options)) (*sns.ListTopicsOutput, error) {
			return &sns.ListTopicsOutput{
				Topics: []types.Topic{{TopicArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-1")}},
			}, nil
		},
		listSubscriptionsFn: func(_ context.Context, _ *sns.ListSubscriptionsInput, _ ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error) {
			return &sns.ListSubscriptionsOutput{
				Subscriptions: []types.Subscription{{SubscriptionArn: ptr("arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa")}},
			}, nil
		},
	}
	svc := getMockedSNSService(mock)
	got := svc.GetAllResources()
	expected := []common.Resource{
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-1", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-1", ResourceType: common.SNS_topic, Region: "eu-west-1"},
		{Arn: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa", TerraformID: "arn:aws:sns:eu-west-1:123456789012:topic-1:sub-aaa", ResourceType: common.SNS_subscription, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}
