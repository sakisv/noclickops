package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/noclickops/common"
)

type S3Client interface {
	ListBuckets(ctx context.Context, params *s3.ListBucketsInput, optFns ...func(*s3.Options)) (*s3.ListBucketsOutput, error)
}

type NoclickopsS3Client struct {
	Client S3Client
	ClientMeta
}

type NoclickopsS3Service struct {
	Clients []NoclickopsS3Client
	common.ServiceMeta
}

func NewS3ServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsS3Service {
	service := NoclickopsS3Service{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsS3Client{
			Client:     s3.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsS3Service) GetAllResources() []common.Resource {
	return s.GetAllBuckets()
}

func (s *NoclickopsS3Service) GetAllBuckets() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var continuationToken *string = nil
		for {
			res, err := rc.Client.ListBuckets(context.TODO(), &s3.ListBucketsInput{
				BucketRegion:      &rc.Region,
				ContinuationToken: continuationToken,
			})
			if err != nil {
				log.Fatal(err)
			}

			for _, el := range res.Buckets {
				resources = append(resources, common.Resource{TerraformID: *el.Name, ResourceType: common.S3_bucket, Region: rc.Region})
			}

			if res.ContinuationToken == nil {
				break
			}
			continuationToken = res.ContinuationToken
		}
	}
	return resources
}
