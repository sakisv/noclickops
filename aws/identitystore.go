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
	Client []IdentityStoreClient
	Meta   common.ClientMeta
}

func NewIdentityStoreClientFromConfigs(cfg []awssdk.Config) NoClickopsIdentityStoreClient {
	clopsClient := NoClickopsIdentityStoreClient{}
	clopsClient.Meta = common.ClientMeta{
		Regional:    false,
		ServiceName: "identitystore",
	}
	for _, cfg := range cfg {
		clopsClient.Client = append(clopsClient.Client, identitystore.NewFromConfig(cfg))
		if clopsClient.Meta.Regional == false {
			break
		}
	}
	return clopsClient
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
