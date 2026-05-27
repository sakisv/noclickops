package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/elasticache"
	"github.com/aws/aws-sdk-go-v2/service/elasticache/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedElasticacheService(mock *mockElasticacheClient) aws.NoclickopsElasticacheService {
	return aws.NoclickopsElasticacheService{
		Clients: []aws.NoclickopsElasticacheClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "elasticache"},
	}
}

func TestGetCacheClusters_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockElasticacheClient{
		describeCacheClustersFn: func(_ context.Context, params *elasticache.DescribeCacheClustersInput, _ ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
			callCount++
			if callCount == 1 {
				return &elasticache.DescribeCacheClustersOutput{
					Marker:        ptr("next"),
					CacheClusters: []types.CacheCluster{{CacheClusterId: ptr("cluster-1"), ARN: ptr("arn:aws:elasticache:eu-west-1:123456789012:cluster:cluster-1")}},
				}, nil
			}
			if params.Marker == nil || *params.Marker != "next" {
				return nil, fmt.Errorf("wrong Marker, expected 'next' got '%v'", params.Marker)
			}
			return &elasticache.DescribeCacheClustersOutput{
				CacheClusters: []types.CacheCluster{{CacheClusterId: ptr("cluster-2"), ARN: ptr("arn:aws:elasticache:eu-west-1:123456789012:cluster:cluster-2")}},
			}, nil
		},
	}
	service := getMockedElasticacheService(mock)
	got := service.GetClusters()
	expected := []common.Resource{
		{Arn: "arn:aws:elasticache:eu-west-1:123456789012:cluster:cluster-1", TerraformID: "cluster-1", ResourceType: common.Elasticache_cluster, Region: "eu-west-1"},
		{Arn: "arn:aws:elasticache:eu-west-1:123456789012:cluster:cluster-2", TerraformID: "cluster-2", ResourceType: common.Elasticache_cluster, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, got)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to DescribeCacheClusters, got %d", callCount)
	}
}

func TestGetCacheClusters_ErrorIsSkipped(t *testing.T) {
	mock := &mockElasticacheClient{
		describeCacheClustersFn: func(_ context.Context, _ *elasticache.DescribeCacheClustersInput, _ ...func(*elasticache.Options)) (*elasticache.DescribeCacheClustersOutput, error) {
			return nil, fmt.Errorf("some error")
		},
	}
	service := getMockedElasticacheService(mock)
	got := service.GetClusters()
	if len(got) != 0 {
		t.Errorf("expected no resources on error, got %v", got)
	}
}
