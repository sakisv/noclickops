package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ecs"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedECSService(mock *mockECSClient) aws.NoclickopsECSService {
	return aws.NoclickopsECSService{
		Clients: []aws.NoclickopsECSClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "ecs"},
	}
}

func TestGetECSClustersAndServices_PaginationFollowed(t *testing.T) {
	listClustersCallCount := 0
	listServicesCallCount := 0
	mock := &mockECSClient{
		listClustersFn: func(_ context.Context, params *ecs.ListClustersInput, _ ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
			listClustersCallCount++
			if listClustersCallCount == 1 {
				return &ecs.ListClustersOutput{
					NextToken:   ptr("next-cluster"),
					ClusterArns: []string{"arn:aws:ecs:eu-west-1:123456789012:cluster/cluster-1"},
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next-cluster" {
				return nil, fmt.Errorf("wrong NextToken, expected 'next-cluster' got '%v'", params.NextToken)
			}
			return &ecs.ListClustersOutput{
				ClusterArns: []string{"arn:aws:ecs:eu-west-1:123456789012:cluster/cluster-2"},
			}, nil
		},
		listServicesFn: func(_ context.Context, params *ecs.ListServicesInput, _ ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
			listServicesCallCount++
			if listServicesCallCount == 1 {
				return &ecs.ListServicesOutput{
					NextToken:   ptr("next-service"),
					ServiceArns: []string{"arn:aws:ecs:eu-west-1:123456789012:service/cluster-1/service-1"},
				}, nil
			}
			if listServicesCallCount == 2 {
				if params.NextToken == nil || *params.NextToken != "next-service" {
					return nil, fmt.Errorf("wrong NextToken, expected 'next-service' got '%v'", params.NextToken)
				}
				return &ecs.ListServicesOutput{
					ServiceArns: []string{"arn:aws:ecs:eu-west-1:123456789012:service/cluster-1/service-2"},
				}, nil
			}
			return &ecs.ListServicesOutput{
				ServiceArns: []string{"arn:aws:ecs:eu-west-1:123456789012:service/cluster-2/service-3"},
			}, nil
		},
	}
	service := getMockedECSService(mock)
	got := service.GetClustersAndServices()
	expected := []common.Resource{
		{Arn: "arn:aws:ecs:eu-west-1:123456789012:cluster/cluster-1", TerraformID: "cluster-1", ResourceType: common.ECS_cluster, Region: "eu-west-1"},
		{Arn: "arn:aws:ecs:eu-west-1:123456789012:service/cluster-1/service-1", TerraformID: "cluster-1/service-1", ResourceType: common.ECS_service, Region: "eu-west-1"},
		{Arn: "arn:aws:ecs:eu-west-1:123456789012:service/cluster-1/service-2", TerraformID: "cluster-1/service-2", ResourceType: common.ECS_service, Region: "eu-west-1"},
		{Arn: "arn:aws:ecs:eu-west-1:123456789012:cluster/cluster-2", TerraformID: "cluster-2", ResourceType: common.ECS_cluster, Region: "eu-west-1"},
		{Arn: "arn:aws:ecs:eu-west-1:123456789012:service/cluster-2/service-3", TerraformID: "cluster-2/service-3", ResourceType: common.ECS_service, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if listClustersCallCount != 2 {
		t.Errorf("expected 2 calls to ListClusters, got %d", listClustersCallCount)
	}
	if listServicesCallCount != 3 {
		t.Errorf("expected 3 calls to ListServices, got %d", listServicesCallCount)
	}
}

func TestGetECSClustersAndServices_ListClustersErrorIsSkipped(t *testing.T) {
	mock := &mockECSClient{
		listClustersFn: func(_ context.Context, _ *ecs.ListClustersInput, _ ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
			return nil, fmt.Errorf("some error")
		},
	}
	service := getMockedECSService(mock)
	got := service.GetClustersAndServices()
	if len(got) != 0 {
		t.Errorf("expected no resources on error, got %v", got)
	}
}

func TestGetECSClustersAndServices_ListServicesErrorIsSkipped(t *testing.T) {
	mock := &mockECSClient{
		listClustersFn: func(_ context.Context, _ *ecs.ListClustersInput, _ ...func(*ecs.Options)) (*ecs.ListClustersOutput, error) {
			return &ecs.ListClustersOutput{
				ClusterArns: []string{"arn:aws:ecs:eu-west-1:123456789012:cluster/cluster-1"},
			}, nil
		},
		listServicesFn: func(_ context.Context, _ *ecs.ListServicesInput, _ ...func(*ecs.Options)) (*ecs.ListServicesOutput, error) {
			return nil, fmt.Errorf("some error")
		},
	}
	service := getMockedECSService(mock)
	got := service.GetClustersAndServices()
	expected := []common.Resource{
		{Arn: "arn:aws:ecs:eu-west-1:123456789012:cluster/cluster-1", TerraformID: "cluster-1", ResourceType: common.ECS_cluster, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
}
