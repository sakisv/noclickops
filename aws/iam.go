package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/noclickops/common"
)

type IAMClient interface {
	ListPolicies(ctx context.Context, params *iam.ListPoliciesInput, optFns ...func(*iam.Options)) (*iam.ListPoliciesOutput, error)
	ListUsers(ctx context.Context, params *iam.ListUsersInput, optFns ...func(*iam.Options)) (*iam.ListUsersOutput, error)
	ListGroups(ctx context.Context, params *iam.ListGroupsInput, optFns ...func(*iam.Options)) (*iam.ListGroupsOutput, error)
}

type NoClickopsIAMClient struct {
	Client []IAMClient
	common.ClientMeta
}

func NewIAMClientFromConfigs(cfg []awssdk.Config, meta common.ClientMeta) NoClickopsIAMClient {
	clopsClient := NoClickopsIAMClient{}
	clopsClient.ClientMeta = meta
	clopsClient.Client = append(clopsClient.Client, iam.NewFromConfig(cfg[0]))
	return clopsClient
}

const MAX_ITEMS int32 = 150

func (clops *NoClickopsIAMClient) GetAllResources() []common.Resource {
	var resources []common.Resource

	resources = append(resources, clops.GetAllIAMUsers()...)
	resources = append(resources, clops.GetAllIAMGroups()...)
	resources = append(resources, clops.GetAllPoliciesArns()...)
	return resources
}

func (clops *NoClickopsIAMClient) GetAllPoliciesArns() []common.Resource {
	var resources []common.Resource
	var marker *string = nil
	client := clops.Client[0]
	for {
		res, err := client.ListPolicies(context.TODO(), &iam.ListPoliciesInput{
			MaxItems: awssdk.Int32(MAX_ITEMS),
			Scope:    types.PolicyScopeTypeLocal,
			Marker:   marker,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Policies {
			resources = append(resources, common.Resource{TerraformID: *el.Arn, ResourceType: common.IAM_policy})
		}

		if !res.IsTruncated {
			break
		}
		marker = res.Marker
	}
	return resources
}

func (clops *NoClickopsIAMClient) GetAllIAMUsers() []common.Resource {
	var resources []common.Resource
	var marker *string = nil
	client := clops.Client[0]
	for {
		res, err := client.ListUsers(context.TODO(), &iam.ListUsersInput{
			MaxItems: awssdk.Int32(MAX_ITEMS),
			Marker:   marker,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Users {
			resources = append(resources, common.Resource{TerraformID: *el.UserName, ResourceType: common.IAM_user})
		}

		if !res.IsTruncated {
			break
		}
		marker = res.Marker
	}
	return resources
}

func (clops *NoClickopsIAMClient) GetAllIAMGroups() []common.Resource {
	var resources []common.Resource
	var marker *string = nil
	client := clops.Client[0]
	for {
		res, err := client.ListGroups(context.TODO(), &iam.ListGroupsInput{
			MaxItems: awssdk.Int32(MAX_ITEMS),
			Marker:   marker,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Groups {
			resources = append(resources, common.Resource{TerraformID: *el.GroupName, ResourceType: common.IAM_group})
		}

		if !res.IsTruncated {
			break
		}
		marker = res.Marker
	}
	return resources
}
