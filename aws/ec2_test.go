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
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "ec2", AccountId: "123456789012"},
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
					SecurityGroups: []types.SecurityGroup{{GroupId: ptr("sg-1"), SecurityGroupArn: ptr("arn:aws:ec2:eu-west-1:123456789012:security-group/sg-1")}},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken, expected 'next' got '%v'", params.NextToken)
			}
			return &ec2.DescribeSecurityGroupsOutput{
				SecurityGroups: []types.SecurityGroup{{GroupId: ptr("sg-2"), SecurityGroupArn: ptr("arn:aws:ec2:eu-west-1:123456789012:security-group/sg-2")}},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllSecurityGroups()
	expected := []common.Resource{
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:security-group/sg-1", TerraformID: "sg-1", ResourceType: common.Security_group, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:security-group/sg-2", TerraformID: "sg-2", ResourceType: common.Security_group, Region: "eu-west-1"},
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
							SecurityGroupRuleId:  ptr("sgr-1"),
							SecurityGroupRuleArn: ptr("arn:aws:ec2:eu-west-1:123456789012:security-group-rule/sgr-1"),
							GroupId:              ptr("sg-aaa"),
							IsEgress:             ptr(false),
							IpProtocol:           ptr("tcp"),
							FromPort:             ptr(int32(80)),
							ToPort:               ptr(int32(80)),
							CidrIpv4:             ptr("10.0.0.0/8"),
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
						SecurityGroupRuleId:  ptr("sgr-2"),
						SecurityGroupRuleArn: ptr("arn:aws:ec2:eu-west-1:123456789012:security-group-rule/sgr-2"),
						GroupId:              ptr("sg-bbb"),
						IsEgress:             ptr(true),
						IpProtocol:           ptr("tcp"),
						FromPort:             ptr(int32(443)),
						ToPort:               ptr(int32(443)),
						CidrIpv6:             ptr("::/0"),
					},
				},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllSecurityGroupRules()
	expected := []common.Resource{
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:security-group-rule/sgr-1", TerraformID: "sg-aaa_ingress_tcp_80_80_10.0.0.0/8", ResourceType: common.Security_group_rule, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:security-group-rule/sgr-2", TerraformID: "sg-bbb_egress_tcp_443_443_::/0", ResourceType: common.Security_group_rule, Region: "eu-west-1"},
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
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:instance/i-123456", TerraformID: "i-123456", ResourceType: common.Instance, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:instance/i-23456789", TerraformID: "i-23456789", ResourceType: common.Instance, Region: "eu-west-1"},
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
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:elastic-ip/eip-12345678", TerraformID: "eip-12345678", ResourceType: common.Eip, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:elastic-ip/eip-23456790", TerraformID: "eip-23456790", ResourceType: common.Eip, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
}

func TestGetAllVPCs_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEC2Client{
		describeVpcsFn: func(_ context.Context, params *ec2.DescribeVpcsInput, _ ...func(*ec2.Options)) (*ec2.DescribeVpcsOutput, error) {
			callCount++
			if callCount == 1 {
				return &ec2.DescribeVpcsOutput{
					NextToken: ptr("next"),
					Vpcs:      []types.Vpc{{VpcId: ptr("vpc-1")}},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next', got %v", params.NextToken)
			}
			return &ec2.DescribeVpcsOutput{
				Vpcs: []types.Vpc{{VpcId: ptr("vpc-2")}},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllVPCs()
	expected := []common.Resource{
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:vpc/vpc-1", TerraformID: "vpc-1", ResourceType: common.VPC, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:vpc/vpc-2", TerraformID: "vpc-2", ResourceType: common.VPC, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeVpcs, got %d", callCount)
	}
}

func TestGetAllInternetGateways_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEC2Client{
		describeInternetGatewaysFn: func(_ context.Context, params *ec2.DescribeInternetGatewaysInput, _ ...func(*ec2.Options)) (*ec2.DescribeInternetGatewaysOutput, error) {
			callCount++
			if callCount == 1 {
				return &ec2.DescribeInternetGatewaysOutput{
					NextToken:        ptr("next"),
					InternetGateways: []types.InternetGateway{{InternetGatewayId: ptr("igw-1")}},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next', got %v", params.NextToken)
			}
			return &ec2.DescribeInternetGatewaysOutput{
				InternetGateways: []types.InternetGateway{{InternetGatewayId: ptr("igw-2")}},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllInternetGateways()
	expected := []common.Resource{
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:internet-gateway/igw-1", TerraformID: "igw-1", ResourceType: common.Internet_gateway, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:internet-gateway/igw-2", TerraformID: "igw-2", ResourceType: common.Internet_gateway, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeInternetGateways, got %d", callCount)
	}
}

func TestGetAllNATGateways_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEC2Client{
		describeNatGatewaysFn: func(_ context.Context, params *ec2.DescribeNatGatewaysInput, _ ...func(*ec2.Options)) (*ec2.DescribeNatGatewaysOutput, error) {
			callCount++
			if callCount == 1 {
				return &ec2.DescribeNatGatewaysOutput{
					NextToken:   ptr("next"),
					NatGateways: []types.NatGateway{{NatGatewayId: ptr("nat-1")}},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next', got %v", params.NextToken)
			}
			return &ec2.DescribeNatGatewaysOutput{
				NatGateways: []types.NatGateway{{NatGatewayId: ptr("nat-2")}},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllNATGateways()
	expected := []common.Resource{
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:natgateway/nat-1", TerraformID: "nat-1", ResourceType: common.NAT_gateway, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:natgateway/nat-2", TerraformID: "nat-2", ResourceType: common.NAT_gateway, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeNatGateways, got %d", callCount)
	}
}

func TestGetAllSubnets_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEC2Client{
		describeSubnetsFn: func(_ context.Context, params *ec2.DescribeSubnetsInput, _ ...func(*ec2.Options)) (*ec2.DescribeSubnetsOutput, error) {
			callCount++
			if callCount == 1 {
				return &ec2.DescribeSubnetsOutput{
					NextToken: ptr("next"),
					Subnets:   []types.Subnet{{SubnetId: ptr("subnet-1"), SubnetArn: ptr("arn:aws:ec2:eu-west-1:123456789012:subnet/subnet-1")}},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next', got %v", params.NextToken)
			}
			return &ec2.DescribeSubnetsOutput{
				Subnets: []types.Subnet{{SubnetId: ptr("subnet-2"), SubnetArn: ptr("arn:aws:ec2:eu-west-1:123456789012:subnet/subnet-2")}},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllSubnets()
	expected := []common.Resource{
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:subnet/subnet-1", TerraformID: "subnet-1", ResourceType: common.Subnet, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:subnet/subnet-2", TerraformID: "subnet-2", ResourceType: common.Subnet, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeSubnets, got %d", callCount)
	}
}

func TestGetAllVPCEndpoints_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEC2Client{
		describeVpcEndpointsFn: func(_ context.Context, params *ec2.DescribeVpcEndpointsInput, _ ...func(*ec2.Options)) (*ec2.DescribeVpcEndpointsOutput, error) {
			callCount++
			if callCount == 1 {
				return &ec2.DescribeVpcEndpointsOutput{
					NextToken:    ptr("next"),
					VpcEndpoints: []types.VpcEndpoint{{VpcEndpointId: ptr("vpce-1")}},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next', got %v", params.NextToken)
			}
			return &ec2.DescribeVpcEndpointsOutput{
				VpcEndpoints: []types.VpcEndpoint{{VpcEndpointId: ptr("vpce-2")}},
			}, nil
		},
	}
	client := getMockedEC2Service(mock)
	got := client.GetAllVPCEndpoints()
	expected := []common.Resource{
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:vpc-endpoint/vpce-1", TerraformID: "vpce-1", ResourceType: common.VPC_endpoint, Region: "eu-west-1"},
		{Arn: "arn:aws:ec2:eu-west-1:123456789012:vpc-endpoint/vpce-2", TerraformID: "vpce-2", ResourceType: common.VPC_endpoint, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeVpcEndpoints, got %d", callCount)
	}
}
