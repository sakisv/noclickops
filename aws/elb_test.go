package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedELBService(mock *mockELBClient) aws.NoclickopsELBService {
	return aws.NoclickopsELBService{
		Clients: []aws.NoclickopsELBClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "elb", AccountId: "123456789012"},
	}
}

func TestGetAllClassicLoadBalancers_BasicCase(t *testing.T) {
	mock := &mockELBClient{
		describeLoadBalancersFn: func(_ context.Context, _ *elasticloadbalancing.DescribeLoadBalancersInput, _ ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error) {
			return &elasticloadbalancing.DescribeLoadBalancersOutput{
				LoadBalancerDescriptions: []types.LoadBalancerDescription{
					{LoadBalancerName: ptr("my-classic-lb-1")},
					{LoadBalancerName: ptr("my-classic-lb-2")},
				},
			}, nil
		},
	}
	svc := getMockedELBService(mock)
	got := svc.GetAllClassicLoadBalancers()
	expected := []common.Resource{
		{Arn: "arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/my-classic-lb-1", TerraformID: "my-classic-lb-1", ResourceType: common.ELB_load_balancer, Region: "eu-west-1"},
		{Arn: "arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/my-classic-lb-2", TerraformID: "my-classic-lb-2", ResourceType: common.ELB_load_balancer, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetAllClassicLoadBalancers_NoLoadBalancers(t *testing.T) {
	mock := &mockELBClient{
		describeLoadBalancersFn: func(_ context.Context, _ *elasticloadbalancing.DescribeLoadBalancersInput, _ ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error) {
			return &elasticloadbalancing.DescribeLoadBalancersOutput{}, nil
		},
	}
	svc := getMockedELBService(mock)
	got := svc.GetAllClassicLoadBalancers()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetAllClassicLoadBalancers_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockELBClient{
		describeLoadBalancersFn: func(_ context.Context, params *elasticloadbalancing.DescribeLoadBalancersInput, _ ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error) {
			callCount++
			if callCount == 1 {
				return &elasticloadbalancing.DescribeLoadBalancersOutput{
					LoadBalancerDescriptions: []types.LoadBalancerDescription{{LoadBalancerName: ptr("my-classic-lb-1")}},
					NextMarker:               ptr("next-lb"),
				}, nil
			}
			if params.Marker == nil || *params.Marker != "next-lb" {
				return nil, fmt.Errorf("wrong Marker: expected 'next-lb', got %v", params.Marker)
			}
			return &elasticloadbalancing.DescribeLoadBalancersOutput{
				LoadBalancerDescriptions: []types.LoadBalancerDescription{{LoadBalancerName: ptr("my-classic-lb-2")}},
			}, nil
		},
	}
	svc := getMockedELBService(mock)
	got := svc.GetAllClassicLoadBalancers()
	expected := []common.Resource{
		{Arn: "arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/my-classic-lb-1", TerraformID: "my-classic-lb-1", ResourceType: common.ELB_load_balancer, Region: "eu-west-1"},
		{Arn: "arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/my-classic-lb-2", TerraformID: "my-classic-lb-2", ResourceType: common.ELB_load_balancer, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeLoadBalancers, got %d", callCount)
	}
}
