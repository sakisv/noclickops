package aws

import (
	"context"
	"fmt"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/noclickops/common"
)

type EKSClient interface {
	ListClusters(ctx context.Context, params *eks.ListClustersInput, optFns ...func(*eks.Options)) (*eks.ListClustersOutput, error)
	ListNodegroups(ctx context.Context, params *eks.ListNodegroupsInput, optFns ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error)
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
	return s.GetEKSClustersAndNodegroups()
}

func (s *NoclickopsEKSService) GetEKSClustersAndNodegroups() []common.Resource {
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
				resources = append(resources, common.Resource{TerraformID: el, ResourceType: common.EKS_cluster, Region: rc.Region})
				var nextNodegroupToken *string = nil

				for {
					res2, err := rc.Client.ListNodegroups(context.TODO(), &eks.ListNodegroupsInput{
						ClusterName: &el,
						NextToken:   nextNodegroupToken,
					})

					if err != nil {
						log.Fatal(err)
					}

					for _, nodegroup := range res2.Nodegroups {
						tfId := fmt.Sprintf("%v:%v", el, nodegroup)
						resources = append(resources, common.Resource{TerraformID: tfId, ResourceType: common.EKS_node_group, Region: rc.Region})
					}

					if res2.NextToken == nil {
						break
					}
					nextNodegroupToken = res2.NextToken
				}
			}

			if res.NextToken == nil {
				break
			}
			nextToken = res.NextToken
		}
	}
	return resources
}
