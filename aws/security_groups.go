package aws

import (
	"context"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/noclickops/common"
)

type EC2Client interface {
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	DescribeSecurityGroupRules(ctx context.Context, params *ec2.DescribeSecurityGroupRulesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupRulesOutput, error)
}

func GetAllSecurityGroups(client EC2Client) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	for {
		res, err := client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
			NextToken: nextToken,
		})
		if err != nil {
			log.Fatal(err)
		}

		for _, el := range res.SecurityGroups {
			resources = append(resources, common.Resource{TerraformID: *el.GroupId, ResourceType: common.EC2_securitygroup})
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	return resources
}

func GetAllSecurityGroupRules(client EC2Client) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	for {
		res, err := client.DescribeSecurityGroupRules(context.TODO(), &ec2.DescribeSecurityGroupRulesInput{
			NextToken: nextToken,
		})
		if err != nil {
			log.Fatal(err)
		}

		for _, el := range res.SecurityGroupRules {
			rule_type := "ingress"
			if *el.IsEgress {
				rule_type = "egress"
			}
			var id_pieces = []string{
				*el.GroupId,
				rule_type,
				*el.IpProtocol,
				strconv.Itoa(int(*el.FromPort)),
				strconv.Itoa(int(*el.ToPort)),
			}
			if el.CidrIpv4 != nil {
				id_pieces = append(id_pieces, *el.CidrIpv4)
			}
			if el.CidrIpv6 != nil {
				id_pieces = append(id_pieces, *el.CidrIpv6)
			}
			id := strings.Join(id_pieces[:], "_")
			resources = append(resources, common.Resource{TerraformID: id, ResourceType: common.EC2_securitygrouprule})
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	return resources
}
