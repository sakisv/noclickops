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
	got := aws.GetAllSecurityGroups(mock)
	expected := []common.Resource{
		{TerraformID: "sg-1", ResourceType: common.EC2_securitygroup},
		{TerraformID: "sg-2", ResourceType: common.EC2_securitygroup},
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
	got := aws.GetAllSecurityGroupRules(mock)
	expected := []common.Resource{
		{TerraformID: "sg-aaa_ingress_tcp_80_80_10.0.0.0/8", ResourceType: common.EC2_securitygrouprule},
		{TerraformID: "sg-bbb_egress_tcp_443_443_::/0", ResourceType: common.EC2_securitygrouprule},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeSecurityGroupRules, got %d", callCount)
	}
}
