package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

type IAMClient interface {
	ListPolicies(ctx context.Context, params *iam.ListPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListPoliciesOutput, error)
}

const MAX_ITEMS int32 = 150

func GetAllPoliciesArns(client IAMClient) []string {
	var ids []string
	var marker string
	for {
		res, err := client.ListPolicies(context.TODO(), &iam.ListPoliciesInput{
			MaxItems: aws.Int32(MAX_ITEMS),
			Scope:    types.PolicyScopeTypeLocal,
			Marker:   &marker,
		})

		if err != nil {
			log.Fatal(err)
		}

		for _, el := range res.Policies {
			ids = append(ids, *el.Arn)
		}

		if !res.IsTruncated {
			break
		}
		marker = *res.Marker
	}
	return ids
}
