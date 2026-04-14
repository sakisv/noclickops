package aws

import (
	"context"
	"log"
	"strconv"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsec2 "github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/noclickops/common"
)

type EC2Client interface {
	DescribeSecurityGroups(ctx context.Context, params *awsec2.DescribeSecurityGroupsInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupsOutput, error)
	DescribeSecurityGroupRules(ctx context.Context, params *awsec2.DescribeSecurityGroupRulesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeSecurityGroupRulesOutput, error)
	DescribeInstances(ctx context.Context, params *awsec2.DescribeInstancesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInstancesOutput, error)
}

type NoclickopsEC2Client struct {
	Client EC2Client
	ClientMeta
}

type NoclickopsEC2Service struct {
	Clients []NoclickopsEC2Client
	common.ServiceMeta
}

func NewEC2ServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsEC2Service {
	service := NoclickopsEC2Service{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsEC2Client{
			Client:     awsec2.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsEC2Service) GetAllResources() []common.Resource {
	var resources []common.Resource
	resources = append(resources, s.GetAllSecurityGroups()...)
	resources = append(resources, s.GetAllSecurityGroupRules()...)
	resources = append(resources, s.GetAllEC2Instances()...)
	return resources
}

func (s *NoclickopsEC2Service) GetAllSecurityGroups() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeSecurityGroups(context.TODO(), &awsec2.DescribeSecurityGroupsInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.SecurityGroups {
				resources = append(resources, common.Resource{TerraformID: *el.GroupId, ResourceType: common.Security_group, Region: rc.Region})
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
			res, err := rc.Client.DescribeSecurityGroupRules(context.TODO(), &awsec2.DescribeSecurityGroupRulesInput{
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
				resources = append(resources, common.Resource{TerraformID: id, ResourceType: common.Security_group_rule, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}

func (s *NoclickopsEC2Service) GetAllEC2Instances() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeInstances(context.TODO(), &awsec2.DescribeInstancesInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, reservation := range res.Reservations {
				for _, instance := range reservation.Instances {
					resources = append(resources, common.Resource{TerraformID: *instance.InstanceId, ResourceType: common.Instance, Region: rc.Region})
				}
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}
