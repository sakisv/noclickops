package aws

import (
	"context"
	"log"
	"strings"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/noclickops/common"
)

type ECSClient interface {
	ListClusters(ctx context.Context, params *ecs.ListClustersInput, optFns ...func(*ecs.Options)) (*ecs.ListClustersOutput, error)
	ListServices(ctx context.Context, params *ecs.ListServicesInput, optFns ...func(*ecs.Options)) (*ecs.ListServicesOutput, error)
}

type NoclickopsECSClient struct {
	Client ECSClient
	ClientMeta
}

type NoclickopsECSService struct {
	Clients []NoclickopsECSClient
	common.ServiceMeta
}

func NewECSServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsECSService {
	service := NoclickopsECSService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsECSClient{
			Client:     ecs.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsECSService) GetAllResources() []common.Resource {
	return s.GetClustersAndServices()
}

func (s *NoclickopsECSService) GetClustersAndServices() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var clusterNextToken *string = nil
		for {
			clusterRes, err := rc.Client.ListClusters(context.TODO(), &ecs.ListClustersInput{
				NextToken: clusterNextToken,
			})
			if err != nil {
				log.Printf("warning: %v", err)
				break
			}

			for _, clusterArn := range clusterRes.ClusterArns {
				// ecs cluster arn format
				// arn:${Partition}:ecs:${Region}:${Account}:cluster/${ClusterName}
				parts := strings.Split(clusterArn, "/")
				clusterName := parts[len(parts)-1]
				resources = append(resources, common.Resource{Arn: clusterArn, TerraformID: clusterName, ResourceType: common.ECS_cluster, Region: rc.Region})

				var nextToken *string = nil
				for {
					res, err := rc.Client.ListServices(context.TODO(), &ecs.ListServicesInput{
						Cluster:   awssdk.String(clusterArn),
						NextToken: nextToken,
					})
					if err != nil {
						log.Printf("warning: %v", err)
						break
					}

					for _, serviceArn := range res.ServiceArns {
						// ecs service arn format
						// arn:${Partition}:ecs:${Region}:${Account}:service/${ClusterName}/${ServiceName}
						//
						// tf import: cluster-name/service-name
						tfId := strings.SplitN(serviceArn, ":service/", 2)[1]
						resources = append(resources, common.Resource{Arn: serviceArn, TerraformID: tfId, ResourceType: common.ECS_service, Region: rc.Region})
					}

					if res.NextToken == nil {
						break
					}
					nextToken = res.NextToken
				}
			}

			if clusterRes.NextToken == nil {
				break
			}
			clusterNextToken = clusterRes.NextToken
		}
	}
	return resources
}
