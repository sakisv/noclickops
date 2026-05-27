package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/noclickops/common"
)

type ElasticacheClient interface {
	DescribeCacheClusters(ctx context.Context, params *elasticache.DescribeCacheClustersInput, optFns ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error)
}

type NoclickopsElasticacheClient struct {
	Client ElasticacheClient
	ClientMeta
}

type NoclickopsElasticacheService struct {
	Clients []NoclickopsElasticacheClient
	common.ServiceMeta
}

func NewElasticacheServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsElasticacheService {
	service := NoclickopsElasticacheService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsElasticacheClient{
			Client:     elasticache.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsElasticacheService) GetAllResources() []common.Resource {
	return s.GetClusters()
}

func (s *NoclickopsElasticacheService) GetClusters() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.DescribeCacheClusters(context.TODO(), &elasticache.DescribeCacheClustersInput{
				Marker:                                  marker,
				ShowCacheClustersNotInReplicationGroups: awssdk.Bool(true),
			})
			if err != nil {
				log.Printf("warning: %v", err)
				break
			}

			for _, cluster := range res.CacheClusters {
				resources = append(resources, common.Resource{Arn: *cluster.ARN, TerraformID: *cluster.CacheClusterId, ResourceType: common.Elasticache_cluster, Region: rc.Region})
			}

			if res.Marker == nil {
				break
			}
			marker = res.Marker
		}
	}
	return resources
}
