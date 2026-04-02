package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/noclickops/common"
)

type EKSClient interface {
	ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error)
}

func GetAllEKSClusters(client EKSClient) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	var include = []string{"all"}

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
