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

type NoclickopsIdentityStoreClient struct {
	Client IdentityStoreClient
	ClientMeta
}

type NoclickopsIdentityStoreService struct {
	Clients        []NoclickopsIdentityStoreClient
	SSOAdminClient *NoclickopsSSOAdminService
	common.ServiceMeta
}

func NewIdentityStoreClientFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta, ssoClient *NoclickopsSSOAdminService) NoclickopsIdentityStoreService {
	service := NoclickopsIdentityStoreService{ServiceMeta: meta, SSOAdminClient: ssoClient}
	service.Clients = append(service.Clients, NoclickopsIdentityStoreClient{
		Client:     identitystore.NewFromConfig(cfg[0]),
		ClientMeta: ClientMeta{Region: cfg[0].Region},
	})
	return service
}

func (s *NoclickopsIdentityStoreService) GetAllResources() []common.Resource {
	var resources []common.Resource
	resources = append(resources, s.GetAllIdentityStoreUsers(s.SSOAdminClient)...)
	resources = append(resources, s.GetAllIdentityStoreGroups(s.SSOAdminClient)...)
	return resources
}

func (s *NoclickopsIdentityStoreService) GetAllIdentityStoreUsers(ssoadmin_client *NoclickopsSSOAdminService) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	client := s.Clients[0].Client

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

func (s *NoclickopsIdentityStoreService) GetAllIdentityStoreGroups(ssoadmin_client *NoclickopsSSOAdminService) []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	client := s.Clients[0].Client

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
