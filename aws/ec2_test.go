package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/ec2/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedEC2Service(mock *mockEC2Client) aws.NoclickopsEC2Service {
	return aws.NoclickopsEC2Service{
		Clients: []aws.NoclickopsEC2Client{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "ec2"},
	}
}

func TestGetAllSecurityGroups_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEC2Client{
		describeSecurityGroupsFn: func(_ context.Context, params *ec2.DescribeSecurityGroupsInput, _ ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error) {
			callCount++
			if callCount == 1 {
				return &ec2.DescribeSecurityGroupsOutput{
					NextToken:      ptr("next"),
					SecurityGroups: []types.SecurityGroup{{GroupId: ptr("sg-1")}},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken, expected 'next' got '%v'", params.NextToken)
			}
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []types.SecurityGroup{{GroupId: ptr("sg-2")}},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllSecurityGroups()
	expected := []common.Resource{
		{TerraformID: "sg-1", ResourceType: common.Security_group, Region: "eu-west-1"},
		{TerraformID: "sg-2", ResourceType: common.Security_group, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeSecurityGroups, got %d", callCount)
	}
}

func TestGetAllSecurityGroupRules_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEC2Client{
		describeSecurityGroupRulesFn: func(_ context.Context, params *ec2.DescribeSecurityGroupRulesInput, _ ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupRulesOutput, error) {
			callCount++
			if callCount == 1 {
				return &ec2.DescribeSecurityGroupRulesOutput{
					NextToken: ptr("next"),
					SecurityGroupRules: []types.SecurityGroupRule{
						{
							SecurityGroupRuleId: ptr("sgr-1"),
							GroupId:             ptr("sg-aaa"),
							IsEgress:            ptr(false),
							IpProtocol:          ptr("tcp"),
							FromPort:            ptr(int32(80)),
							ToPort:              ptr(int32(80)),
							CidrIpv4:            ptr("10.0.0.0/8"),
						},
					},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken, expected 'next' got '%v'", params.NextToken)
			}
			return &ec2.DescribeSecurityGroupRulesOutput{
				SecurityGroupRules: []types.SecurityGroupRule{
					{
						SecurityGroupRuleId: ptr("sgr-2"),
						GroupId:             ptr("sg-bbb"),
						IsEgress:            ptr(true),
						IpProtocol:          ptr("tcp"),
						FromPort:            ptr(int32(443)),
						ToPort:              ptr(int32(443)),
						CidrIpv6:            ptr("::/0"),
					},
				},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllSecurityGroupRules()
	expected := []common.Resource{
		{TerraformID: "sg-aaa_ingress_tcp_80_80_10.0.0.0/8", ResourceType: common.Security_group_rule, Region: "eu-west-1"},
		{TerraformID: "sg-bbb_egress_tcp_443_443_::/0", ResourceType: common.Security_group_rule, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeSecurityGroupRules, got %d", callCount)
	}
}

func TestGetAllInstances_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEC2Client{
		describeInstancesFn: func(_ context.Context, params *ec2.DescribeInstancesInput, _ ...func(*ec2.Options)) (*ec2.DescribeInstancesOutput, error) {
			callCount++
			if callCount == 1 {
				return &ec2.DescribeInstancesOutput{
					NextToken: ptr("next"),
					Reservations: []types.Reservation{
						{
							Instances: []types.Instance{
								{InstanceId: ptr("i-123456")},
							},
						},
					},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken, expected 'next' got '%v'", params.NextToken)
			}
			return &ec2.DescribeInstancesOutput{
				Reservations: []types.Reservation{
					{
						Instances: []types.Instance{
							{InstanceId: ptr("i-23456789")},
						},
					},
				},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllEC2Instances()
	expected := []common.Resource{
		{TerraformID: "i-123456", ResourceType: common.Instance, Region: "eu-west-1"},
		{TerraformID: "i-23456789", ResourceType: common.Instance, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeInstances, got %d", callCount)
	}
}

func TestGetAllAddresses(t *testing.T) {
	mock := &mockEC2Client{
		describeAddressesFn: func(_ context.Context, params *ec2.DescribeAddressesInput, _ ...func(*ec2.Options)) (*ec2.DescribeAddressesOutput, error) {
			return &ec2.DescribeAddressesOutput{
				Addresses: []types.Address{
					{AllocationId: ptr("eip-12345678")},
					{AllocationId: ptr("eip-23456790")},
				},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllElasticIPs()
	expected := []common.Resource{
		{TerraformID: "eip-12345678", ResourceType: common.Eip, Region: "eu-west-1"},
		{TerraformID: "eip-23456790", ResourceType: common.Eip, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
}
