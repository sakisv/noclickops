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

type NoClickopsSSMRegionalClient struct {
	Client SSMClient
	ClientMeta
}

type NoClickopsSSMService struct {
	Clients []NoClickopsSSMRegionalClient
	common.ServiceMeta
}

func NewSSMClientFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoClickopsSSMService {
	service := NoClickopsSSMService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoClickopsSSMRegionalClient{
			Client:     ssm.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoClickopsSSMService) GetAllResources() []common.Resource {
	return s.GetAllParametersNames()
}

func (s *NoClickopsSSMService) GetAllParametersNames() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var nextToken string
		for {
			res, err := rc.Client.GetParametersByPath(context.TODO(), &ssm.GetParametersByPathInput{
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
	}
	return resources
}
