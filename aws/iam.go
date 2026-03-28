package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
)

func GetAllPoliciesArns(cfg aws.Config) []string {
	client := iam.NewFromConfig(cfg)

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
	//fmt.Println("Found " + strconv.Itoa(len(ids)) + " policies")
	return ids
}
