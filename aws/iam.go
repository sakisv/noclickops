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

func GetAllPoliciesArns(client IAMClient) []string {
	res, err := client.ListPolicies(context.TODO(), &iam.ListPoliciesInput{
		MaxItems: aws.Int32(500),
		Scope:    types.PolicyScopeTypeLocal,
	})

	if err != nil {
		log.Fatal(err)
	}

	var ids []string
	for _, el := range res.Policies {
		ids = append(ids, *el.Arn)
	}
	return ids
}
