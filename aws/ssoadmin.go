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

type NoClickopsSSOAdminClient struct {
	Client []SSOAdminClient
	common.ServiceMeta
}

func NewSSOAdminClientFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoClickopsSSOAdminClient {
	clopsClient := NoClickopsSSOAdminClient{}
	clopsClient.ServiceMeta = meta
	for _, cfg := range cfg {
		clopsClient.Client = append(clopsClient.Client, ssoadmin.NewFromConfig(cfg))
	}
	return clopsClient
}

func (clops *NoClickopsSSOAdminClient) GetAllResources() []common.Resource {
	return clops.GetAllPermissionSets()
}

func (clops *NoClickopsSSOAdminClient) getSSOInstanceId() string {
	instances := clops.GetAllSSOInstances()
	if len(instances) != 1 {
		println("Found more than 1 SSO Instances, returning")
		return ""
	}

	return *instances[0].IdentityStoreId
}

func (clops *NoClickopsSSOAdminClient) getSSOInstanceArn() string {
	instances := clops.GetAllSSOInstances()
	if len(instances) != 1 {
		println("Found more than 1 SSO Instances, returning")
		return ""
	}

	return *instances[0].InstanceArn
}

func (clops *NoClickopsSSOAdminClient) GetAllSSOInstances() []types.InstanceMetadata {
	var resources []types.InstanceMetadata
	var nextToken *string = nil
	client := clops.Client[0]
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

func (clops *NoClickopsSSOAdminClient) GetAllPermissionSets() []common.Resource {
	var resources []common.Resource
	var nextToken *string = nil
	client := clops.Client[0]

	instance_arn := clops.getSSOInstanceArn()
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
