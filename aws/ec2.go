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

type NoclickopsEC2Client struct {
	Client EC2Client
	ClientMeta
}

type NoclickopsEC2Service struct {
	Clients []NoclickopsEC2Client
	common.ServiceMeta
}

func NewEC2ClientFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsEC2Service {
	service := NoclickopsEC2Service{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsEC2Client{
			Client:     ec2.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsEC2Service) GetAllResources() []common.Resource {
	var resources []common.Resource
	resources = append(resources, s.GetAllSecurityGroups()...)
	resources = append(resources, s.GetAllSecurityGroupRules()...)
	return resources
}

func (s *NoclickopsEC2Service) GetAllSecurityGroups() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeSecurityGroups(context.TODO(), &ec2.DescribeSecurityGroupsInput{
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
	}
	return resources
}

func (s *NoclickopsEC2Service) GetAllSecurityGroupRules() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeSecurityGroupRules(context.TODO(), &ec2.DescribeSecurityGroupRulesInput{
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
	}
	return resources
}
