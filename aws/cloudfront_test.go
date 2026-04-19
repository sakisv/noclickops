package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/cloudfront"
	"github.com/aws/aws-sdk-go-v2/service/cloudfront/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedCloudFrontService(mock *mockCloudFrontClient) aws.NoclickopsCloudFrontService {
	return aws.NoclickopsCloudFrontService{
		Clients: []aws.NoclickopsCloudFrontClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "us-east-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: true, ServiceName: "cloudfront"},
	}
}

func TestGetAllDistributions_BasicCase(t *testing.T) {
	mock := &mockCloudFrontClient{
		listDistributionsFn: func(_ context.Context, _ *cloudfront.ListDistributionsInput, _ ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error) {
			return &cloudfront.ListDistributionsOutput{
				DistributionList: &types.DistributionList{
					IsTruncated: ptr(false),
					Items: []types.DistributionSummary{
						{Id: ptr("E1EXAMPLE1"), ARN: ptr("arn:aws:cloudfront::123456789012:distribution/E1EXAMPLE1")},
						{Id: ptr("E2EXAMPLE2"), ARN: ptr("arn:aws:cloudfront::123456789012:distribution/E2EXAMPLE2")},
					},
				},
			}, nil
		},
	}
	svc := getMockedCloudFrontService(mock)
	got := svc.GetAllDistributions()
	expected := []common.Resource{
		{Arn: "arn:aws:cloudfront::123456789012:distribution/E1EXAMPLE1", TerraformID: "E1EXAMPLE1", ResourceType: common.CloudFront_distribution, Region: "us-east-1"},
		{Arn: "arn:aws:cloudfront::123456789012:distribution/E2EXAMPLE2", TerraformID: "E2EXAMPLE2", ResourceType: common.CloudFront_distribution, Region: "us-east-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetAllDistributions_NoDistributions(t *testing.T) {
	mock := &mockCloudFrontClient{
		listDistributionsFn: func(_ context.Context, _ *cloudfront.ListDistributionsInput, _ ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error) {
			return &cloudfront.ListDistributionsOutput{
				DistributionList: &types.DistributionList{
					IsTruncated: ptr(false),
				},
			}, nil
		},
	}
	svc := getMockedCloudFrontService(mock)
	got := svc.GetAllDistributions()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetAllDistributions_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockCloudFrontClient{
		listDistributionsFn: func(_ context.Context, params *cloudfront.ListDistributionsInput, _ ...func(*cloudfront.Options)) (*cloudfront.ListDistributionsOutput, error) {
			callCount++
			if callCount == 1 {
				return &cloudfront.ListDistributionsOutput{
					DistributionList: &types.DistributionList{
						IsTruncated: ptr(true),
						NextMarker:  ptr("next-dist"),
						Items:       []types.DistributionSummary{{Id: ptr("E1EXAMPLE1"), ARN: ptr("arn:aws:cloudfront::123456789012:distribution/E1EXAMPLE1")}},
					},
				}, nil
			}
			if params.Marker == nil || *params.Marker != "next-dist" {
				return nil, fmt.Errorf("wrong Marker: expected 'next-dist', got %v", params.Marker)
			}
			return &cloudfront.ListDistributionsOutput{
				DistributionList: &types.DistributionList{
					IsTruncated: ptr(false),
					Items:       []types.DistributionSummary{{Id: ptr("E2EXAMPLE2"), ARN: ptr("arn:aws:cloudfront::123456789012:distribution/E2EXAMPLE2")}},
				},
			}, nil
		},
	}
	svc := getMockedCloudFrontService(mock)
	got := svc.GetAllDistributions()
	expected := []common.Resource{
		{Arn: "arn:aws:cloudfront::123456789012:distribution/E1EXAMPLE1", TerraformID: "E1EXAMPLE1", ResourceType: common.CloudFront_distribution, Region: "us-east-1"},
		{Arn: "arn:aws:cloudfront::123456789012:distribution/E2EXAMPLE2", TerraformID: "E2EXAMPLE2", ResourceType: common.CloudFront_distribution, Region: "us-east-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListDistributions, got %d", callCount)
	}
}
