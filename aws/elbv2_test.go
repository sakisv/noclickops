package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedELBV2Service(mock *mockELBV2Client) aws.NoclickopsELBV2Service {
	return aws.NoclickopsELBV2Service{
		Clients: []aws.NoclickopsELBV2Client{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "elbv2"},
	}
}

func TestGetAllLoadBalancers_BasicCase(t *testing.T) {
	mock := &mockELBV2Client{
		describeLoadBalancersFn: func(_ context.Context, _ *elasticloadbalancingv2.DescribeLoadBalancersInput, _ ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
			return &elasticloadbalancingv2.DescribeLoadBalancersOutput{
				LoadBalancers: []types.LoadBalancer{
					{LoadBalancerArn: ptr("arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/app/my-alb/abc123")},
					{LoadBalancerArn: ptr("arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/net/my-nlb/def456")},
				},
			}, nil
		},
	}
	svc := getMockedELBV2Service(mock)
	got := svc.GetAllLoadBalancers()
	expected := []common.Resource{
		{TerraformID: "arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/app/my-alb/abc123", ResourceType: common.ELBV2_load_balancer, Region: "eu-west-1"},
		{TerraformID: "arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/net/my-nlb/def456", ResourceType: common.ELBV2_load_balancer, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetAllLoadBalancers_NoLoadBalancers(t *testing.T) {
	mock := &mockELBV2Client{
		describeLoadBalancersFn: func(_ context.Context, _ *elasticloadbalancingv2.DescribeLoadBalancersInput, _ ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
			return &elasticloadbalancingv2.DescribeLoadBalancersOutput{}, nil
		},
	}
	svc := getMockedELBV2Service(mock)
	got := svc.GetAllLoadBalancers()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetAllLoadBalancers_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockELBV2Client{
		describeLoadBalancersFn: func(_ context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, _ ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error) {
			callCount++
			if callCount == 1 {
				return &elasticloadbalancingv2.DescribeLoadBalancersOutput{
					LoadBalancers: []types.LoadBalancer{
						{LoadBalancerArn: ptr("arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/app/my-alb/abc123")},
					},
					NextMarker: ptr("next-lb"),
				}, nil
			}
			if params.Marker == nil || *params.Marker != "next-lb" {
				return nil, fmt.Errorf("wrong Marker: expected 'next-lb', got %v", params.Marker)
			}
			return &elasticloadbalancingv2.DescribeLoadBalancersOutput{
				LoadBalancers: []types.LoadBalancer{
					{LoadBalancerArn: ptr("arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/net/my-nlb/def456")},
				},
			}, nil
		},
	}
	svc := getMockedELBV2Service(mock)
	got := svc.GetAllLoadBalancers()
	expected := []common.Resource{
		{TerraformID: "arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/app/my-alb/abc123", ResourceType: common.ELBV2_load_balancer, Region: "eu-west-1"},
		{TerraformID: "arn:aws:elasticloadbalancing:eu-west-1:123456789012:loadbalancer/net/my-nlb/def456", ResourceType: common.ELBV2_load_balancer, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeLoadBalancers, got %d", callCount)
	}
}
