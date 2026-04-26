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

type NoclickopsSSOAdminClient struct {
	Client SSOAdminClient
	ClientMeta
}

type NoclickopsSSOAdminService struct {
	Clients []NoclickopsSSOAdminClient
	common.ServiceMeta
}

func NewSSOAdminServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsSSOAdminService {
	service := NoclickopsSSOAdminService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsSSOAdminClient{
			Client:     ssoadmin.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsSSOAdminService) GetAllResources() []common.Resource {
	return s.GetAllPermissionSets()
}

func (s *NoclickopsSSOAdminService) getSSOInstanceId() string {
	instances := s.GetAllSSOInstances()
	if len(instances) == 1 {
		return *instances[0].IdentityStoreId
	}
	return ""
}

func (s *NoclickopsSSOAdminService) getSSOInstanceArn() string {
	instances := s.GetAllSSOInstances()
	if len(instances) == 1 {
		return *instances[0].InstanceArn
	}
	return ""
}

func (s *NoclickopsSSOAdminService) GetAllSSOInstances() []types.InstanceMetadata {
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

func (s *NoclickopsSSOAdminService) GetAllPermissionSets() []common.Resource {
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
			resources = append(resources, common.Resource{Arn: el, TerraformID: tf_id, ResourceType: common.SSOAdmin_permission_set})
		}

		if res.NextToken == nil {
			break
		}
		nextToken = res.NextToken
	}
	return resources
}
