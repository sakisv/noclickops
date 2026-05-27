package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatchlogs"
	"github.com/noclickops/common"
)

type CloudwatchLogsClient interface {
	ListLogGroups(ctx context.Context, params *cloudwatchlogs.ListLogGroupsInput, optFns ...func(*cloudwatchlogs.Options)) (*cloudwatchlogs.ListLogGroupsOutput, error)
}

type NoclickopsCloudwatchLogsClient struct {
	Client CloudwatchLogsClient
	ClientMeta
}

type NoclickopsCloudwatchLogsService struct {
	Clients []NoclickopsCloudwatchLogsClient
	common.ServiceMeta
}

func NewCloudwatchLogsServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsCloudwatchLogsService {
	service := NoclickopsCloudwatchLogsService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsCloudwatchLogsClient{
			Client:     cloudwatchlogs.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsCloudwatchLogsService) GetAllResources() []common.Resource {
	return s.GetLogGroups()
}

func (s *NoclickopsCloudwatchLogsService) GetLogGroups() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		for {
			res, err := rc.Client.ListLogGroups(context.TODO(), &cloudwatchlogs.ListLogGroupsInput{
				NextToken: nextToken,
			})
			if err != nil {
				log.Printf("warning: %v", err)
				break
			}

			for _, loggroup := range res.LogGroups {
				resources = append(resources, common.Resource{Arn: *loggroup.LogGroupArn, TerraformID: *loggroup.LogGroupName, ResourceType: common.Cloudwatchlogs_log_group, Region: rc.Region})
			}


			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}
