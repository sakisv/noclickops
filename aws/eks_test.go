package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedEKSService(mock *mockEKSClient) aws.NoclickopsEKSService {
	return aws.NoclickopsEKSService{
		Clients: []aws.NoclickopsEKSClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "eks"},
	}
}

func TestGetEKSClustersAndNodegroups_BasicCase(t *testing.T) {
	mock := &mockEKSClient{
		listClustersFn: func(_ context.Context, _ *eks.ListClustersInput, _ ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{
				Clusters: []string{"cluster-1"},
			}, nil
		},
		listNodegroupsFn: func(_ context.Context, _ *eks.ListNodegroupsInput, _ ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
			return &eks.ListNodegroupsOutput{
				Nodegroups: []string{"ng-1", "ng-2"},
			}, nil
		},
	}
	svc := getMockedEKSService(mock)
	got := svc.GetEKSClustersAndNodegroups()
	expected := []common.Resource{
		{TerraformID: "cluster-1", ResourceType: common.EKS_cluster, Region: "eu-west-1"},
		{TerraformID: "cluster-1:ng-1", ResourceType: common.EKS_node_group, Region: "eu-west-1"},
		{TerraformID: "cluster-1:ng-2", ResourceType: common.EKS_node_group, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetEKSClustersAndNodegroups_NoClusters(t *testing.T) {
	mock := &mockEKSClient{
		listClustersFn: func(_ context.Context, _ *eks.ListClustersInput, _ ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{}, nil
		},
		listNodegroupsFn: func(_ context.Context, _ *eks.ListNodegroupsInput, _ ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
			t.Error("ListNodegroups should not be called when there are no clusters")
			return &eks.ListNodegroupsOutput{}, nil
		},
	}
	svc := getMockedEKSService(mock)
	got := svc.GetEKSClustersAndNodegroups()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetEKSClustersAndNodegroups_NoNodegroups(t *testing.T) {
	mock := &mockEKSClient{
		listClustersFn: func(_ context.Context, _ *eks.ListClustersInput, _ ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{
				Clusters: []string{"cluster-1"},
			}, nil
		},
		listNodegroupsFn: func(_ context.Context, _ *eks.ListNodegroupsInput, _ ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
			return &eks.ListNodegroupsOutput{}, nil
		},
	}
	svc := getMockedEKSService(mock)
	got := svc.GetEKSClustersAndNodegroups()
	expected := []common.Resource{
		{TerraformID: "cluster-1", ResourceType: common.EKS_cluster, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetEKSClustersAndNodegroups_ClusterPaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockEKSClient{
		listClustersFn: func(_ context.Context, params *eks.ListClustersInput, _ ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			callCount++
			if callCount == 1 {
				return &eks.ListClustersOutput{
					Clusters:  []string{"cluster-1"},
					NextToken: ptr("next-cluster"),
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next-cluster" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next-cluster', got %v", params.NextToken)
			}
			return &eks.ListClustersOutput{
				Clusters: []string{"cluster-2"},
			}, nil
		},
		listNodegroupsFn: func(_ context.Context, _ *eks.ListNodegroupsInput, _ ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
			return &eks.ListNodegroupsOutput{}, nil
		},
	}
	svc := getMockedEKSService(mock)
	got := svc.GetEKSClustersAndNodegroups()
	expected := []common.Resource{
		{TerraformID: "cluster-1", ResourceType: common.EKS_cluster, Region: "eu-west-1"},
		{TerraformID: "cluster-2", ResourceType: common.EKS_cluster, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListClusters, got %d", callCount)
	}
}

func TestGetEKSClustersAndNodegroups_NodegroupPaginationFollowed(t *testing.T) {
	ngCallCount := 0
	mock := &mockEKSClient{
		listClustersFn: func(_ context.Context, _ *eks.ListClustersInput, _ ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{
				Clusters: []string{"cluster-1"},
			}, nil
		},
		listNodegroupsFn: func(_ context.Context, params *eks.ListNodegroupsInput, _ ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
			ngCallCount++
			if ngCallCount == 1 {
				return &eks.ListNodegroupsOutput{
					Nodegroups: []string{"ng-1"},
					NextToken:  ptr("next-ng"),
				}, nil
			}
			if params.NextToken == nil || *params.NextToken != "next-ng" {
				return nil, fmt.Errorf("wrong NextToken: expected 'next-ng', got %v", params.NextToken)
			}
			return &eks.ListNodegroupsOutput{
				Nodegroups: []string{"ng-2"},
			}, nil
		},
	}
	svc := getMockedEKSService(mock)
	got := svc.GetEKSClustersAndNodegroups()
	expected := []common.Resource{
		{TerraformID: "cluster-1", ResourceType: common.EKS_cluster, Region: "eu-west-1"},
		{TerraformID: "cluster-1:ng-1", ResourceType: common.EKS_node_group, Region: "eu-west-1"},
		{TerraformID: "cluster-1:ng-2", ResourceType: common.EKS_node_group, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if ngCallCount != 2 {
		t.Errorf("expected 2 calls to ListNodegroups, got %d", ngCallCount)
	}
}

func TestGetEKSClustersAndNodegroups_MultipleClustersSeparateNodegroups(t *testing.T) {
	mock := &mockEKSClient{
		listClustersFn: func(_ context.Context, _ *eks.ListClustersInput, _ ...func(*eks.Options)) (*eks.ListClustersOutput, error) {
			return &eks.ListClustersOutput{
				Clusters: []string{"cluster-a", "cluster-b"},
			}, nil
		},
		listNodegroupsFn: func(_ context.Context, params *eks.ListNodegroupsInput, _ ...func(*eks.Options)) (*eks.ListNodegroupsOutput, error) {
			if *params.ClusterName == "cluster-a" {
				return &eks.ListNodegroupsOutput{Nodegroups: []string{"ng-a1"}}, nil
			}
			return &eks.ListNodegroupsOutput{Nodegroups: []string{"ng-b1"}}, nil
		},
	}
	svc := getMockedEKSService(mock)
	got := svc.GetEKSClustersAndNodegroups()
	expected := []common.Resource{
		{TerraformID: "cluster-a", ResourceType: common.EKS_cluster, Region: "eu-west-1"},
		{TerraformID: "cluster-a:ng-a1", ResourceType: common.EKS_node_group, Region: "eu-west-1"},
		{TerraformID: "cluster-b", ResourceType: common.EKS_cluster, Region: "eu-west-1"},
		{TerraformID: "cluster-b:ng-b1", ResourceType: common.EKS_node_group, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}
