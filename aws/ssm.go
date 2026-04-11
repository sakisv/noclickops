package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/noclickops/common"
)

type SSMClient interface {
	GetParametersByPath(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error)
}

type NoClickopsSSMClient struct {
	Client []SSMClient
	common.ServiceMeta
}

func NewSSMClientFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoClickopsSSMClient {
	clopsClient := NoClickopsSSMClient{}
	clopsClient.ServiceMeta = meta
	for _, cfg := range cfg {
		clopsClient.Client = append(clopsClient.Client, ssm.NewFromConfig(cfg))
	}
	return clopsClient
}

func (clops *NoClickopsSSMClient) GetAllResources() []common.Resource {
	return clops.GetAllParametersNames()
}

func (clops *NoClickopsSSMClient) GetAllParametersNames() []common.Resource {
	var resources []common.Resource
	var nextToken string
	client := clops.Client[0]
	for {
		res, err := client.GetParametersByPath(context.TODO(), &ssm.GetParametersByPathInput{
			Path:       awssdk.String("/"),
			MaxResults: awssdk.Int32(10),
			Recursive:  awssdk.Bool(true),
			NextToken:  awssdk.String(nextToken),
		})

		if err != nil {
			log.Fatal(err)
		}
		for _, el := range res.Parameters {
			resources = append(resources, common.Resource{TerraformID: *el.Name, ResourceType: common.SSM_parameter})
		}

		if res.NextToken == nil || *res.NextToken == "" {
			break
		}
		nextToken = *res.NextToken
	}
	return resources
}
