package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	awsautoscaling "github.com/aws/aws-sdk-go-v2/service/autoscaling"
	"github.com/noclickops/common"
)

type AutoscalingClient interface {
	DescribeAutoScalingGroups(ctx context.Context, params *awsautoscaling.DescribeAutoScalingGroupsInput, optFns ...func(*awsautoscaling.Options)) (*awsautoscaling.DescribeAutoScalingGroupsOutput, error)
}

type NoclickopsAutoscalingClient struct {
	Client AutoscalingClient
	ClientMeta
}

type NoclickopsAutoscalingService struct {
	Clients []NoclickopsAutoscalingClient
	common.ServiceMeta
}

func NewAutoscalingServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsAutoscalingService {
	service := NoclickopsAutoscalingService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsAutoscalingClient{
			Client:     awsautoscaling.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsAutoscalingService) GetAllResources() []common.Resource {
	return s.GetAllAutoScalingGroups()
}

func (s *NoclickopsAutoscalingService) GetAllAutoScalingGroups() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.DescribeAutoScalingGroups(context.TODO(), &awsautoscaling.DescribeAutoScalingGroupsInput{
				NextToken:        nextToken,
				IncludeInstances: awssdk.Bool(false),
			})
			if err != nil {
				log.Printf("warning: %v", err)
				break
			}

			for _, el := range res.AutoScalingGroups {
				resources = append(resources, common.Resource{Arn: *el.AutoScalingGroupARN, TerraformID: *el.AutoScalingGroupName, ResourceType: common.Autoscaling_group, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}
