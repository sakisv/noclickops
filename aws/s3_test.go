package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedS3Service(mock *mockS3Client) aws.NoclickopsS3Service {
	return aws.NoclickopsS3Service{
		Clients: []aws.NoclickopsS3Client{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "s3"},
	}
}

func TestGetAllBuckets_BasicCase(t *testing.T) {
	mock := &mockS3Client{
		listBucketsFn: func(_ context.Context, _ *s3.ListBucketsInput, _ ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			return &s3.ListBucketsOutput{
				Buckets: []types.Bucket{
					{Name: ptr("bucket-1")},
					{Name: ptr("bucket-2")},
				},
			}, nil
		},
	}
	svc := getMockedS3Service(mock)
	got := svc.GetAllBuckets()
	expected := []common.Resource{
		{Arn: "arn:aws:s3:::bucket-1", TerraformID: "bucket-1", ResourceType: common.S3_bucket, Region: "eu-west-1"},
		{Arn: "arn:aws:s3:::bucket-2", TerraformID: "bucket-2", ResourceType: common.S3_bucket, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetAllBuckets_NoBuckets(t *testing.T) {
	mock := &mockS3Client{
		listBucketsFn: func(_ context.Context, _ *s3.ListBucketsInput, _ ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			return &s3.ListBucketsOutput{}, nil
		},
	}
	svc := getMockedS3Service(mock)
	got := svc.GetAllBuckets()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetAllBuckets_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockS3Client{
		listBucketsFn: func(_ context.Context, params *s3.ListBucketsInput, _ ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			callCount++
			if callCount == 1 {
				return &s3.ListBucketsOutput{
					Buckets:           []types.Bucket{{Name: ptr("bucket-1")}},
					ContinuationToken: ptr("next-bucket"),
				}, nil
			}
			if params.ContinuationToken == nil || *params.ContinuationToken != "next-bucket" {
				return nil, fmt.Errorf("wrong ContinuationToken: expected 'next-bucket', got %v", params.ContinuationToken)
			}
			return &s3.ListBucketsOutput{
				Buckets: []types.Bucket{{Name: ptr("bucket-2")}},
			}, nil
		},
	}
	svc := getMockedS3Service(mock)
	got := svc.GetAllBuckets()
	expected := []common.Resource{
		{Arn: "arn:aws:s3:::bucket-1", TerraformID: "bucket-1", ResourceType: common.S3_bucket, Region: "eu-west-1"},
		{Arn: "arn:aws:s3:::bucket-2", TerraformID: "bucket-2", ResourceType: common.S3_bucket, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListBuckets, got %d", callCount)
	}
}

func TestGetAllBuckets_RegionFilterPassedToAPI(t *testing.T) {
	mock := &mockS3Client{
		listBucketsFn: func(_ context.Context, params *s3.ListBucketsInput, _ ...func(*s3.Options)) (*s3.ListBucketsOutput, error) {
			if params.BucketRegion == nil || *params.BucketRegion != "eu-west-1" {
				return nil, fmt.Errorf("expected BucketRegion 'eu-west-1', got %v", params.BucketRegion)
			}
			return &s3.ListBucketsOutput{}, nil
		},
	}
	svc := getMockedS3Service(mock)
	svc.GetAllBuckets()
}
