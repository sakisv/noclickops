package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/noclickops/common"
)

func getSSOInstanceId(client SSOAdminClient) string {
	instances := GetAllSSOInstances(client)
	if len(instances) != 1 {
		println("Found more than 1 SSO Instances, returning")
		return ""
	}

	return instances[0].TerraformID
}

type IdentityStoreClient interface {
	ListUsers(ctx context.Context, params *identitystore.ListUsersInput, optFns ...func(*identitystore.Options)) (*identitystore.ListUsersOutput, error)
	ListGroups(ctx context.Context, params *identitystore.ListGroupsInput, optFns ...func(*identitystore.Options)) (*identitystore.ListGroupsOutput, error)
}

func GetAllIdentityStoreUsers(client IdentityStoreClient, ssoadmin_client SSOAdminClient) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil

	instance_id := getSSOInstanceId(ssoadmin_client)
	if instance_id == "" {
		return resources
	}
	for {
		res, err := client.ListUsers(context.TODO(), &identitystore.ListUsersInput{
			IdentityStoreId: &instance_id,
			NextToken:       nextToken,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Users {
			tf_id := fmt.Sprintf("%v/%v", instance_id, *el.UserId)
			resources = append(resources, common.Resource{TerraformID: tf_id, ResourceType: common.IdentityStore_user})
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	return resources
}

func GetAllIdentityStoreGroups(client IdentityStoreClient, ssoadmin_client SSOAdminClient) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil

	instance_id := getSSOInstanceId(ssoadmin_client)
	if instance_id == "" {
		return resources
	}
	for {
		res, err := client.ListGroups(context.TODO(), &identitystore.ListGroupsInput{
			IdentityStoreId: &instance_id,
			NextToken:       nextToken,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Groups {
			tf_id := fmt.Sprintf("%v/%v", instance_id, *el.GroupId)
			resources = append(resources, common.Resource{TerraformID: tf_id, ResourceType: common.IdentityStore_group})
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}

	return resources
}
