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
	DescribeAddresses(ctx context.Context, params *awsec2.DescribeAddressesInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeAddressesOutput, error)
	DescribeVpcs(ctx context.Context, params *awsec2.DescribeVpcsInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeVpcsOutput, error)
	DescribeInternetGateways(ctx context.Context, params *awsec2.DescribeInternetGatewaysInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeInternetGatewaysOutput, error)
	DescribeNatGateways(ctx context.Context, params *awsec2.DescribeNatGatewaysInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeNatGatewaysOutput, error)
	DescribeSubnets(ctx context.Context, params *awsec2.DescribeSubnetsInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeSubnetsOutput, error)
	DescribeVpcEndpoints(ctx context.Context, params *awsec2.DescribeVpcEndpointsInput, optFns ...func(*awsec2.Options)) (*awsec2.DescribeVpcEndpointsOutput, error)
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
	resources = append(resources, s.GetAllElasticIPs()...)
	resources = append(resources, s.GetAllVPCs()...)
	resources = append(resources, s.GetAllInternetGateways()...)
	resources = append(resources, s.GetAllNATGateways()...)
	resources = append(resources, s.GetAllSubnets()...)
	resources = append(resources, s.GetAllVPCEndpoints()...)
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

func (s *NoclickopsEC2Service) GetAllElasticIPs() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		res, err := rc.Client.DescribeAddresses(context.TODO(), &awsec2.DescribeAddressesInput{})
		if err != nil {
			log.Fatal(err)
		}

		for _, address := range res.Addresses {
			resources = append(resources, common.Resource{TerraformID: *address.AllocationId, ResourceType: common.Eip, Region: rc.Region})
		}
	}
	return resources
}

func (s *NoclickopsEC2Service) GetAllVPCs() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeVpcs(context.TODO(), &awsec2.DescribeVpcsInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.Vpcs {
				resources = append(resources, common.Resource{TerraformID: *el.VpcId, ResourceType: common.VPC, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}

func (s *NoclickopsEC2Service) GetAllInternetGateways() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeInternetGateways(context.TODO(), &awsec2.DescribeInternetGatewaysInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.InternetGateways {
				resources = append(resources, common.Resource{TerraformID: *el.InternetGatewayId, ResourceType: common.Internet_gateway, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}

func (s *NoclickopsEC2Service) GetAllNATGateways() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeNatGateways(context.TODO(), &awsec2.DescribeNatGatewaysInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.NatGateways {
				resources = append(resources, common.Resource{TerraformID: *el.NatGatewayId, ResourceType: common.NAT_gateway, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}

func (s *NoclickopsEC2Service) GetAllSubnets() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeSubnets(context.TODO(), &awsec2.DescribeSubnetsInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.Subnets {
				resources = append(resources, common.Resource{TerraformID: *el.SubnetId, ResourceType: common.Subnet, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}

func (s *NoclickopsEC2Service) GetAllVPCEndpoints() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeVpcEndpoints(context.TODO(), &awsec2.DescribeVpcEndpointsInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.VpcEndpoints {
				resources = append(resources, common.Resource{TerraformID: *el.VpcEndpointId, ResourceType: common.VPC_endpoint, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}
