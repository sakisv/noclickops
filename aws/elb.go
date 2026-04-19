package aws

import (
	"context"
	"fmt"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancing"
	"github.com/noclickops/common"
)

type ELBClient interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancing.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancing.Options)) (*elasticloadbalancing.DescribeLoadBalancersOutput, error)
}

type NoclickopsELBClient struct {
	Client ELBClient
	ClientMeta
}

type NoclickopsELBService struct {
	Clients []NoclickopsELBClient
	common.ServiceMeta
}

func NewELBServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsELBService {
	service := NoclickopsELBService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsELBClient{
			Client:     elasticloadbalancing.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsELBService) GetAllResources() []common.Resource {
	return s.GetAllClassicLoadBalancers()
}

func (s *NoclickopsELBService) GetAllClassicLoadBalancers() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancing.DescribeLoadBalancersInput{
				Marker: marker,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.LoadBalancerDescriptions {
				// clb arn format:
				//  arn:${Partition}:elasticloadbalancing:${Region}:${Account}:loadbalancer/${LoadBalancerName}
				arn := fmt.Sprintf("arn:aws:elasticloadbalancing:%v:%v:loadbalancer/%v", rc.Region, s.AccountId, *el.LoadBalancerName)
				resources = append(resources, common.Resource{Arn: arn, TerraformID: *el.LoadBalancerName, ResourceType: common.ELB_load_balancer, Region: rc.Region})
			}

			if res.NextMarker == nil {
				break
			}
			marker = res.NextMarker
		}
	}
	return resources
}
