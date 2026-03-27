package aws

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func GetAllPoliciesArns(cfg aws.Config) []string {
	iam_client := iam.NewFromConfig(cfg)

	res_iam, err := iam_client.ListPolicies(context.TODO(), &iam.ListPoliciesInput{
		MaxItems: aws.Int32(500),
		Scope:    types.PolicyScopeTypeLocal,
	})

	if err != nil {
		log.Fatal(err)
	}

	var ids []string
	for _, el := range res_iam.Policies {
		ids = append(ids, *el.Arn)
	}
	fmt.Println("Found " + strconv.Itoa(len(ids)) + " policies")
	return ids
}
