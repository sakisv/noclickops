package aws

import (
	"context"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/noclickops/common"
)

type SSOAdminClient interface {
	ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error)
}

func GetAllSSOInstances(client SSOAdminClient) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	for {
		res, err := client.ListInstances(context.TODO(), &ssoadmin.ListInstancesInput{
			NextToken: nextToken,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Instances {
			resources = append(resources, common.Resource{TerraformID: *el.IdentityStoreId, ResourceType: common.SSOAdmin_identitystoreinstance})
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	return resources
}
