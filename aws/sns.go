package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/sns"
	"github.com/noclickops/common"
)

type SNSClient interface {
	ListTopics(ctx context.Context, params *sns.ListTopicsInput, optFns ...func(*sns.Options)) (*sns.ListTopicsOutput, error)
	ListSubscriptions(ctx context.Context, params *sns.ListSubscriptionsInput, optFns ...func(*sns.Options)) (*sns.ListSubscriptionsOutput, error)
}

type NoclickopsSNSClient struct {
	Client SNSClient
	ClientMeta
}

type NoclickopsSNSService struct {
	Clients []NoclickopsSNSClient
	common.ServiceMeta
}

func NewSNSServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsSNSService {
	service := NoclickopsSNSService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsSNSClient{
			Client:     sns.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsSNSService) GetAllResources() []common.Resource {
	var resources []common.Resource
	resources = append(resources, s.GetAllTopics()...)
	resources = append(resources, s.GetAllSubscriptions()...)

	return resources
}

func (s *NoclickopsSNSService) GetAllTopics() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.ListTopics(context.TODO(), &sns.ListTopicsInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.Topics {
				resources = append(resources, common.Resource{Arn: *el.TopicArn, TerraformID: *el.TopicArn, ResourceType: common.SNS_topic, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}

func (s *NoclickopsSNSService) GetAllSubscriptions() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.ListSubscriptions(context.TODO(), &sns.ListSubscriptionsInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.Subscriptions {
				resources = append(resources, common.Resource{Arn: *el.SubscriptionArn, TerraformID: *el.SubscriptionArn, ResourceType: common.SNS_subscription, Region: rc.Region})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}
