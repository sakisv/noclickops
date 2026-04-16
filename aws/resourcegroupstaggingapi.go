package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	"github.com/noclickops/common"
)

type ResourceGroupTaggingAPIClient interface {
	GetResources(ctx context.Context, params *resourcegroupstaggingapi.GetResourcesInput, optFns ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error)
}

type NoclickopsResourceGroupTaggingAPIClient struct {
	Client ResourceGroupTaggingAPIClient
	ClientMeta
}

type NoclickopsResourceGroupTaggingAPIService struct {
	Clients []NoclickopsResourceGroupTaggingAPIClient
	common.ServiceMeta
}

func NewResourceGroupTaggingAPIServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsResourceGroupTaggingAPIService {
	service := NoclickopsResourceGroupTaggingAPIService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsResourceGroupTaggingAPIClient{
			Client:     resourcegroupstaggingapi.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsResourceGroupTaggingAPIService) GetAllResources() []common.Resource {
	return []common.Resource{}
}

func (s *NoclickopsResourceGroupTaggingAPIService) GetResourcesWithTags(key string, values []string) []common.Resource {
	var resources []common.Resource

	for _, rc := range s.Clients {
		var paginationToken *string = nil
		for {
			resp, err := rc.Client.GetResources(context.TODO(), &resourcegroupstaggingapi.GetResourcesInput{
				PaginationToken: paginationToken,
				TagFilters: []types.TagFilter{
					{
						Key:    awssdk.String(key),
						Values: values,
					},
				},
			})

			if err != nil {
				log.Fatal(err)
			}

			for _, tagMapping := range resp.ResourceTagMappingList {
				resources = append(resources, common.Resource{TerraformID: *tagMapping.ResourceARN, Region: rc.Region})
			}

			if resp.PaginationToken == nil || *resp.PaginationToken == "" {
				break
			}

			paginationToken = resp.PaginationToken
		}

	}

	return resources
}
