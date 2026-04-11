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

type NoclickopsSSMClient struct {
	Client SSMClient
	ClientMeta
}

type NoclickopsSSMService struct {
	Clients []NoclickopsSSMClient
	common.ServiceMeta
}

func NewSSMServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsSSMService {
	service := NoclickopsSSMService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsSSMClient{
			Client:     ssm.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsSSMService) GetAllResources() []common.Resource {
	return s.GetAllParametersNames()
}

func (s *NoclickopsSSMService) GetAllParametersNames() []common.Resource {
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
				resources = append(resources, common.Resource{TerraformID: *el.Name, ResourceType: common.SSM_parameter, Region: rc.Region})
			}

			if res.NextToken == nil || *res.NextToken == "" {
				break
			}
			nextToken = *res.NextToken
		}
	}
	return resources
}
