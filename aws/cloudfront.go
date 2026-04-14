package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/noclickops/common"
)

type CloudFrontClient interface {
	ListDistributions(ctx context.Context, params *cloudfront.ListDistributionsInput, optFns ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error)
}

type NoclickopsCloudFrontClient struct {
	Client CloudFrontClient
	ClientMeta
}

type NoclickopsCloudFrontService struct {
	Clients []NoclickopsCloudFrontClient
	common.ServiceMeta
}

func NewCloudFrontServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsCloudFrontService {
	service := NoclickopsCloudFrontService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsCloudFrontClient{
			Client:     cloudfront.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: "global"},
		})
	}
	return service
}

func (s *NoclickopsCloudFrontService) GetAllResources() []common.Resource {
	return s.GetAllDistributions()
}

func (s *NoclickopsCloudFrontService) GetAllDistributions() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.ListDistributions(context.TODO(), &cloudfront.ListDistributionsInput{
				Marker: marker,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.DistributionList.Items {
				resources = append(resources, common.Resource{TerraformID: *el.Id, ResourceType: common.CloudFront_distribution, Region: rc.Region})
			}

			if !*res.DistributionList.IsTruncated {
				break
			}
			marker = res.DistributionList.NextMarker
		}
	}
	return resources
}
