package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticloadbalancingv2"
	"github.com/noclickops/common"
)

type ELBV2Client interface {
	DescribeLoadBalancers(ctx context.Context, params *elasticloadbalancingv2.DescribeLoadBalancersInput, optFns ...func(*elasticloadbalancingv2.Options)) (*elasticloadbalancingv2.DescribeLoadBalancersOutput, error)
}

type NoclickopsELBV2Client struct {
	Client ELBV2Client
	ClientMeta
}

type NoclickopsELBV2Service struct {
	Clients []NoclickopsELBV2Client
	common.ServiceMeta
}

func NewELBV2ServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsELBV2Service {
	service := NoclickopsELBV2Service{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsELBV2Client{
			Client:     elasticloadbalancingv2.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsELBV2Service) GetAllResources() []common.Resource {
	return s.GetAllLoadBalancers()
}

func (s *NoclickopsELBV2Service) GetAllLoadBalancers() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.DescribeLoadBalancers(context.TODO(), &elasticloadbalancingv2.DescribeLoadBalancersInput{
				Marker: marker,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.LoadBalancers {
				resources = append(resources, common.Resource{TerraformID: *el.LoadBalancerArn, ResourceType: common.ELBV2_load_balancer, Region: rc.Region})
			}

			if res.NextMarker == nil {
				break
			}
			marker = res.NextMarker
		}
	}
	return resources
}
