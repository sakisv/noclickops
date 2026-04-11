package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/noclickops/common"
)

type EKSClient interface {
	ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error)
}

type NoclickopsEKSClient struct {
	Client EKSClient
	ClientMeta
}

type NoclickopsEKSService struct {
	Clients []NoclickopsEKSClient
	common.ServiceMeta
}

func NewEKSServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsEKSService {
	service := NoclickopsEKSService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsEKSClient{
			Client:     eks.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsEKSService) GetAllResources() []common.Resource {
	return s.GetAllEKSClusters()
}

func (s *NoclickopsEKSService) GetAllEKSClusters() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken *string = nil
		var include = []string{"all"}
		for {
			res, err := rc.Client.ListClusters(context.TODO(), &eks.ListClustersInput{
				Include:   include,
				NextToken: nextToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.Clusters {
				resources = append(resources, common.Resource{TerraformID: el, ResourceType: common.EKS_cluster})
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}
