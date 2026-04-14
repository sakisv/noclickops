package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/noclickops/common"
)

type RDSClient interface {
	DescribeDBInstances(ctx context.Context, params *rds.DescribeDBInstancesInput, optFns ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error)
	DescribeDBClusters(ctx context.Context, params *rds.DescribeDBClustersInput, optFns ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error)
}

type NoclickopsRDSClient struct {
	Client RDSClient
	ClientMeta
}

type NoclickopsRDSService struct {
	Clients []NoclickopsRDSClient
	common.ServiceMeta
}

func NewRDSServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsRDSService {
	service := NoclickopsRDSService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsRDSClient{
			Client:     rds.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsRDSService) GetAllResources() []common.Resource {
	var resources []common.Resource
	resources = append(resources, s.GetAllDBInstances()...)
	resources = append(resources, s.GetAllDBClusters()...)
	return resources
}

func (s *NoclickopsRDSService) GetAllDBInstances() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.DescribeDBInstances(context.TODO(), &rds.DescribeDBInstancesInput{
				Marker: marker,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.DBInstances {
				resources = append(resources, common.Resource{TerraformID: *el.DBInstanceIdentifier, ResourceType: common.DB_instance, Region: rc.Region})
			}

			if res.Marker == nil {
				break
			}
			marker = res.Marker
		}
	}
	return resources
}

func (s *NoclickopsRDSService) GetAllDBClusters() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.DescribeDBClusters(context.TODO(), &rds.DescribeDBClustersInput{
				Marker: marker,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.DBClusters {
				resources = append(resources, common.Resource{TerraformID: *el.DBClusterIdentifier, ResourceType: common.RDS_cluster, Region: rc.Region})
			}

			if res.Marker == nil {
				break
			}
			marker = res.Marker
		}
	}
	return resources
}
