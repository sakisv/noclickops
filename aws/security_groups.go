package aws

import (
	"context"
	"log"
	"strconv"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/noclickops/common"
)

type EC2Client interface {
	DescribeSecurityGroups(ctx context.Context, params *ec2.DescribeSecurityGroupsInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupsOutput, error)
	DescribeSecurityGroupRules(ctx context.Context, params *ec2.DescribeSecurityGroupRulesInput, optFns ...func(*ec2.Options)) (*ec2.DescribeSecurityGroupRulesOutput, error)
}

type NoClickopsEC2Client struct {
	Client []EC2Client
	Meta   common.ClientMeta
}

func NewEC2ClientFromConfigs(cfg []awssdk.Config) NoClickopsEC2Client {
	clopsClient := NoClickopsEC2Client{}
	clopsClient.Meta = common.ClientMeta{
		Regional:    true,
		ServiceName: "ec2",
	}
	for _, cfg := range cfg {
		clopsClient.Client = append(clopsClient.Client, ec2.NewFromConfig(cfg))
		if clopsClient.Meta.Regional == false {
			break
		}
	}
	return clopsClient
}

func (clops *NoClickopsEC2Client) GetAllSecurityGroups() []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	client := clops.Client[0]
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

func (clops *NoClickopsEC2Client) GetAllSecurityGroupRules() []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	client := clops.Client[0]
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
