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

type NoClickopsEKSClient struct {
	Client []EKSClient
	common.ServiceMeta
}

func NewEKSClientFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoClickopsEKSClient {
	clopsClient := NoClickopsEKSClient{}
	clopsClient.ServiceMeta = meta
	for _, cfg := range cfg {
		clopsClient.Client = append(clopsClient.Client, eks.NewFromConfig(cfg))
	}
	return clopsClient
}

func (clops *NoClickopsEKSClient) GetAllResources() []common.Resource {
	return clops.GetAllEKSClusters()
}

func (clops *NoClickopsEKSClient) GetAllEKSClusters() []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	var include = []string{"all"}
	client := clops.Client[0]

	for {
		res, err := client.ListClusters(context.TODO(), &eks.ListClustersInput{
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
	return resources
}
