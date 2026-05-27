package aws

import (
	"context"
	"fmt"
	"log"

	awssdk "github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/noclickops/common"
)

type DynamodbClient interface {
	ListTables(ctx context.Context, params *dynamodb.ListTablesInput, optFns ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error)
}

type NoclickopsDynamodbClient struct {
	Client DynamodbClient
	ClientMeta
}

type NoclickopsDynamodbService struct {
	Clients []NoclickopsDynamodbClient
	common.ServiceMeta
}

func NewDynamodbServiceFromConfigs(cfg []awssdk.Config, meta common.ServiceMeta) NoclickopsDynamodbService {
	service := NoclickopsDynamodbService{ServiceMeta: meta}
	for _, c := range cfg {
		service.Clients = append(service.Clients, NoclickopsDynamodbClient{
			Client:     dynamodb.NewFromConfig(c),
			ClientMeta: ClientMeta{Region: c.Region},
		})
	}
	return service
}

func (s *NoclickopsDynamodbService) GetAllResources() []common.Resource {
	return s.GetTables()
}

func (s *NoclickopsDynamodbService) GetTables() []common.Resource {
	var resources []common.Resource
	for _, rc := range s.Clients {
		var exclusiveStartTableName *string = nil
		for {
			res, err := rc.Client.ListTables(context.TODO(), &dynamodb.ListTablesInput{
				ExclusiveStartTableName: exclusiveStartTableName,
			})
			if err != nil {
				log.Printf("warning: %v", err)
				break
			}

			for _, table := range res.TableNames {
				// dynamodb table ARN format:
				//  arn:${Partition}:dynamodb:${Region}:${Account}:table/${TableName}
				arn := fmt.Sprintf("arn:aws:dynamodb:%v:%v:table/%v", rc.Region, s.AccountId, table)
				resources = append(resources, common.Resource{Arn: arn, TerraformID: table, ResourceType: common.Dynamodb_table, Region: rc.Region})
			}

			if res.LastEvaluatedTableName == nil {
				break
			}
			exclusiveStartTableName = res.LastEvaluatedTableName
		}
	}
	return resources
}
