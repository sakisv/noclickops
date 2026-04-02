package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/noclickops/common"
)

type SSOAdminClient interface {
	ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error)
}

func getSSOInstanceId(client SSOAdminClient) string {
	instances := GetAllSSOInstances(client)
	if len(instances) != 1 {
		println("Found more than 1 SSO Instances, returning")
		return ""
	}

	return *instances[0].IdentityStoreId
}

func getSSOInstanceArn(client SSOAdminClient) string {
	instances := GetAllSSOInstances(client)
	if len(instances) != 1 {
		println("Found more than 1 SSO Instances, returning")
		return ""
	}

	return *instances[0].InstanceArn
}

func GetAllSSOInstances(client SSOAdminClient) []types.InstanceMetadata {
	var resources []types.InstanceMetadata
	var nextToken *string = nil
	for {
		res, err := client.ListInstances(context.TODO(), &ssoadmin.ListInstancesInput{
			NextToken: nextToken,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Instances {
			resources = append(resources, el)
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	return resources
}
