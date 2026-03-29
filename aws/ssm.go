package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

type SSMClient interface {
	GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
}

func GetAllParametersNames(client SSMClient) []string {
	var ids []string
	var nextToken string
	for {
		res, err := client.GetParametersByPath(context.TODO(), &ssm.GetParametersByPathInput{
			Path:       aws.String("/"),
			MaxResults: aws.Int32(10),
			Recursive:  aws.Bool(true),
			NextToken:  aws.String(nextToken),
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Parameters {
			ids = append(ids, *el.Name)
		}

		if res.NextToken == nil || *res.NextToken == "" {
			break
		}
		nextToken = *res.NextToken
	}
	return ids
}
