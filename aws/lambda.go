package aws

import (
	"context"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/lambda"
	"github.com/noclickops/common"
)

type LambdaClient interface {
	ListFunctions(ctx context.Context, params *lambda.ListFunctionsInput, optFns ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error)
}

type NoclickopsLambdaClient struct {
	Client LambdaClient
	ClientMeta
}

type NoclickopsLambdaService struct {
	Clients []NoclickopsLambdaClient
	common.ServiceMeta
}

func NewLambdaServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsLambdaService {
	service := NoclickopsLambdaService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsLambdaClient{
			Client:     lambda.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsLambdaService) GetAllResources() []common.Resource {
	return s.GetAllLambdaFunctions()
}

func (s *NoclickopsLambdaService) GetAllLambdaFunctions() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var marker *string = nil
		for {
			res, err := rc.Client.ListFunctions(context.TODO(), &lambda.ListFunctionsInput{
				Marker:   marker,
				MaxItems: awssdk.Int32(50),
			})
			if err != nil {
				log.Printf("warning: %v", err)
				break
			}

			for _, el := range res.Functions {
				resources = append(resources, common.Resource{Arn: *el.FunctionArn, TerraformID: *el.FunctionName, ResourceType: common.Lambda_function, Region: rc.Region})
			}

			if res.NextMarker == nil {
				break
			}
			marker = res.NextMarker
		}
	}
	return resources
}
