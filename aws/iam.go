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

type NoclickopsIAMClient struct {
	Client IAMClient
	ClientMeta
}

type NoclickopsIAMService struct {
	Clients []NoclickopsIAMClient
	common.ServiceMeta
}

func NewIAMServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsIAMService {
	service := NoclickopsIAMService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsIAMClient{
			Client:     iam.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: "global"},
		})
	}
	return service
}

const MAX_ITEMS int32 = 150

func (s *NoclickopsIAMService) GetAllResources() []common.Resource {
	var resources []common.Resource
	resources = append(resources, s.GetAllIAMUsers()...)
	resources = append(resources, s.GetAllIAMGroups()...)
	resources = append(resources, s.GetAllPoliciesArns()...)
	return resources
}

func (s *NoclickopsIAMService) GetAllPoliciesArns() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.ListPolicies(context.TODO(), &iam.ListPoliciesInput{
				MaxItems: awssdk.Int32(MAX_ITEMS),
				Scope:    types.PolicyScopeTypeLocal,
				Marker:   marker,
			})

			if err != nil {
				log.Printf("warning: %v", err)
				break
			}
			for _, el := range res.Policies {
				resources = append(resources, common.Resource{Arn: *el.Arn, TerraformID: *el.Arn, ResourceType: common.IAM_policy, Region: rc.Region})
			}

			if !res.IsTruncated {
				break
			}
			marker = res.Marker
		}
	}
	return resources
}

func (s *NoclickopsIAMService) GetAllIAMUsers() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.ListUsers(context.TODO(), &iam.ListUsersInput{
				MaxItems: awssdk.Int32(MAX_ITEMS),
				Marker:   marker,
			})

			if err != nil {
				log.Printf("warning: %v", err)
				break
			}
			for _, el := range res.Users {
				resources = append(resources, common.Resource{Arn: *el.Arn, TerraformID: *el.UserName, ResourceType: common.IAM_user, Region: rc.Region})
			}

			if !res.IsTruncated {
				break
			}
			marker = res.Marker
		}
	}
	return resources
}

func (s *NoclickopsIAMService) GetAllIAMGroups() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.ListGroups(context.TODO(), &iam.ListGroupsInput{
				MaxItems: awssdk.Int32(MAX_ITEMS),
				Marker:   marker,
			})

			if err != nil {
				log.Printf("warning: %v", err)
				break
			}
			for _, el := range res.Groups {
				resources = append(resources, common.Resource{Arn: *el.Arn, TerraformID: *el.GroupName, ResourceType: common.IAM_group, Region: rc.Region})
			}

			if !res.IsTruncated {
				break
			}
			marker = res.Marker
		}
	}
	return resources
}
