package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/noclickops/common"
)

type IAMClient interface {
	ListPolicies(ctx context.Context, params *iam.ListPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListPoliciesOutput, error)
}

const MAX_ITEMS int32 = 150

func GetAllPoliciesArns(client IAMClient) []common.Resource {
	var resources []common.Resource
	var marker *string = nil
	for {
		res, err := client.ListPolicies(context.TODO(), &iam.ListPoliciesInput{
			MaxItems: aws.Int32(MAX_ITEMS),
			Scope:    types.PolicyScopeTypeLocal,
			Marker:   marker,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Policies {
			resources = append(resources, common.Resource{TerraformID: *el.Arn, ResourceType: common.IAM_policy})
		}

		if !res.IsTruncated {
			break
		}
		marker = res.Marker
	}
	return resources
}
