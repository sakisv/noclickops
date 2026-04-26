package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/rds"
	"github.com/aws/aws-sdk-go-v2/service/rds/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedRDSService(mock *mockRDSClient) aws.NoclickopsRDSService {
	return aws.NoclickopsRDSService{
		Clients: []aws.NoclickopsRDSClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "rds"},
	}
}

func TestGetAllDBInstances_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockRDSClient{
		describeDBInstancesFn: func(_ context.Context, params *rds.DescribeDBInstancesInput, _ ...func(*rds.Options)) (*rds.DescribeDBInstancesOutput, error) {
			callCount++
			if callCount == 1 {
				return &rds.DescribeDBInstancesOutput{
					Marker:      ptr("next"),
					DBInstances: []types.DBInstance{{DBInstanceIdentifier: ptr("db-1"), DBInstanceArn: ptr("arn:aws:rds:eu-west-1:123456789012:db:db-1")}},
				}, nil
			}
			if params.Marker == nil || *params.Marker != "next" {
				return nil, fmt.Errorf("wrong Marker, expected 'next' got '%v'", params.Marker)
			}
			return &rds.DescribeDBInstancesOutput{
				DBInstances: []types.DBInstance{{DBInstanceIdentifier: ptr("db-2"), DBInstanceArn: ptr("arn:aws:rds:eu-west-1:123456789012:db:db-2")}},
			}, nil
		},
	}
	client := getMockedRDSService(mock)
	got := client.GetAllDBInstances()
	expected := []common.Resource{
		{Arn: "arn:aws:rds:eu-west-1:123456789012:db:db-1", TerraformID: "db-1", ResourceType: common.DB_instance, Region: "eu-west-1"},
		{Arn: "arn:aws:rds:eu-west-1:123456789012:db:db-2", TerraformID: "db-2", ResourceType: common.DB_instance, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeDBInstances, got %d", callCount)
	}
}

func TestGetAllDBClusters_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockRDSClient{
		describeDBClustersFn: func(_ context.Context, params *rds.DescribeDBClustersInput, _ ...func(*rds.Options)) (*rds.DescribeDBClustersOutput, error) {
			callCount++
			if callCount == 1 {
				return &rds.DescribeDBClustersOutput{
					Marker:     ptr("next"),
					DBClusters: []types.DBCluster{{DBClusterIdentifier: ptr("cluster-1"), DBClusterArn: ptr("arn:aws:rds:eu-west-1:123456789012:cluster:cluster-1")}},
				}, nil
			}
			if params.Marker == nil || *params.Marker != "next" {
				return nil, fmt.Errorf("wrong Marker, expected 'next' got '%v'", params.Marker)
			}
			return &rds.DescribeDBClustersOutput{
				DBClusters: []types.DBCluster{{DBClusterIdentifier: ptr("cluster-2"), DBClusterArn: ptr("arn:aws:rds:eu-west-1:123456789012:cluster:cluster-2")}},
			}, nil
		},
	}
	client := getMockedRDSService(mock)
	got := client.GetAllDBClusters()
	expected := []common.Resource{
		{Arn: "arn:aws:rds:eu-west-1:123456789012:cluster:cluster-1", TerraformID: "cluster-1", ResourceType: common.RDS_cluster, Region: "eu-west-1"},
		{Arn: "arn:aws:rds:eu-west-1:123456789012:cluster:cluster-2", TerraformID: "cluster-2", ResourceType: common.RDS_cluster, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeDBClusters, got %d", callCount)
	}
}
