package aws

import (
	"context"
	"fmt"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/noclickops/common"
)

type IdentityStoreClient interface {
	ListUsers(ctx context.Context, params *identitystore.ListUsersInput, optFns ...func(*identitystore.Options)) (*identitystore.ListUsersOutput, error)
	ListGroups(ctx context.Context, params *identitystore.ListGroupsInput, optFns ...func(*identitystore.Options)) (*identitystore.ListGroupsOutput, error)
}

type NoClickopsIdentityStoreClient struct {
	Client         []IdentityStoreClient
	SSOAdminClient *NoClickopsSSOAdminClient
	common.ClientMeta
}

func NewIdentityStoreClientFromConfigs(cfg []awssdk.Config, meta common.ClientMeta, ssoClient *NoClickopsSSOAdminClient) NoClickopsIdentityStoreClient {
	clopsClient := NoClickopsIdentityStoreClient{}
	clopsClient.ClientMeta = meta
	clopsClient.SSOAdminClient = ssoClient
	clopsClient.Client = append(clopsClient.Client, identitystore.NewFromConfig(cfg[0]))
	return clopsClient
}

func (clops *NoClickopsIdentityStoreClient) GetAllResources() []common.Resource {
	var resources []common.Resource
	resources = append(resources, clops.GetAllIdentityStoreUsers(clops.SSOAdminClient)...)
	resources = append(resources, clops.GetAllIdentityStoreGroups(clops.SSOAdminClient)...)
	return resources
}

func (clops *NoClickopsIdentityStoreClient) GetAllIdentityStoreUsers(ssoadmin_client *NoClickopsSSOAdminClient) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	client := clops.Client[0]

	instance_id := ssoadmin_client.getSSOInstanceId()
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

func (clops *NoClickopsIdentityStoreClient) GetAllIdentityStoreGroups(ssoadmin_client *NoClickopsSSOAdminClient) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	client := clops.Client[0]

	instance_id := ssoadmin_client.getSSOInstanceId()
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
