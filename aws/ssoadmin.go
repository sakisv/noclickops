package aws

import (
	"context"
	"fmt"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin/types"
	"github.com/noclickops/common"
)

type SSOAdminClient interface {
	ListInstances(ctx context.Context, params *ssoadmin.ListInstancesInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListInstancesOutput, error)
	ListPermissionSets(ctx context.Context, params *ssoadmin.ListPermissionSetsInput, optFns ...func(*ssoadmin.Options)) (*ssoadmin.ListPermissionSetsOutput, error)
}

type NoClickopsSSOAdminRegionalClient struct {
	Client SSOAdminClient
	ClientMeta
}

type NoClickopsSSOAdminService struct {
	Clients []NoClickopsSSOAdminRegionalClient
	common.ServiceMeta
}

func NewSSOAdminClientFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoClickopsSSOAdminService {
	service := NoClickopsSSOAdminService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoClickopsSSOAdminRegionalClient{
			Client:     ssoadmin.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoClickopsSSOAdminService) GetAllResources() []common.Resource {
	return s.GetAllPermissionSets()
}

func (s *NoClickopsSSOAdminService) getSSOInstanceId() string {
	instances := s.GetAllSSOInstances()
	if len(instances) == 1 {
		return *instances[0].IdentityStoreId
	}
	return ""
}

func (s *NoClickopsSSOAdminService) getSSOInstanceArn() string {
	instances := s.GetAllSSOInstances()
	if len(instances) == 1 {
		return *instances[0].InstanceArn
	}
	return ""
}

func (s *NoClickopsSSOAdminService) GetAllSSOInstances() []types.InstanceMetadata {
	var resources []types.InstanceMetadata
	var nextToken *string = nil
	client := s.Clients[0].Client
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

func (s *NoClickopsSSOAdminService) GetAllPermissionSets() []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	client := s.Clients[0].Client

	instance_arn := s.getSSOInstanceArn()
	if instance_arn == "" {
		return resources
	}

	for {
		res, err := client.ListPermissionSets(context.TODO(), &ssoadmin.ListPermissionSetsInput{
			InstanceArn: &instance_arn,
			NextToken:   nextToken,
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.PermissionSets {
			tf_id := fmt.Sprintf("%v,%v", el, instance_arn)
			resources = append(resources, common.Resource{TerraformID: tf_id, ResourceType: common.SSOAdmin_permissionset})
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}
	return resources
}
