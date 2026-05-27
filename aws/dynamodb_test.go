package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedDynamodbService(mock *mockDynamodbClient) aws.NoclickopsDynamodbService {
	return aws.NoclickopsDynamodbService{
		Clients: []aws.NoclickopsDynamodbClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "dynamodb", AccountId: "123456789012"},
	}
}

func TestGetDynamodbTables_BasicCase(t *testing.T) {
	mock := &mockDynamodbClient{
		listTablesFn: func(_ context.Context, _ *dynamodb.ListTablesInput, _ ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
			return &dynamodb.ListTablesOutput{
				TableNames: []string{"table-1", "table-2"},
			}, nil
		},
	}
	svc := getMockedDynamodbService(mock)
	got := svc.GetTables()
	expected := []common.Resource{
		{Arn: "arn:aws:dynamodb:eu-west-1:123456789012:table/table-1", TerraformID: "table-1", ResourceType: common.Dynamodb_table, Region: "eu-west-1"},
		{Arn: "arn:aws:dynamodb:eu-west-1:123456789012:table/table-2", TerraformID: "table-2", ResourceType: common.Dynamodb_table, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
}

func TestGetDynamodbTables_NoTables(t *testing.T) {
	mock := &mockDynamodbClient{
		listTablesFn: func(_ context.Context, _ *dynamodb.ListTablesInput, _ ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
			return &dynamodb.ListTablesOutput{}, nil
		},
	}
	svc := getMockedDynamodbService(mock)
	got := svc.GetTables()
	if len(got) != 0 {
		t.Errorf("expected empty result, got %v", got)
	}
}

func TestGetDynamodbTables_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockDynamodbClient{
		listTablesFn: func(_ context.Context, params *dynamodb.ListTablesInput, _ ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
			callCount++
			if callCount == 1 {
				return &dynamodb.ListTablesOutput{
					TableNames:            []string{"table-1"},
					LastEvaluatedTableName: ptr("table-1"),
				}, nil
			}
			if params.ExclusiveStartTableName == nil || *params.ExclusiveStartTableName != "table-1" {
				return nil, fmt.Errorf("wrong ExclusiveStartTableName: expected 'table-1', got %v", params.ExclusiveStartTableName)
			}
			return &dynamodb.ListTablesOutput{
				TableNames: []string{"table-2"},
			}, nil
		},
	}
	svc := getMockedDynamodbService(mock)
	got := svc.GetTables()
	expected := []common.Resource{
		{Arn: "arn:aws:dynamodb:eu-west-1:123456789012:table/table-1", TerraformID: "table-1", ResourceType: common.Dynamodb_table, Region: "eu-west-1"},
		{Arn: "arn:aws:dynamodb:eu-west-1:123456789012:table/table-2", TerraformID: "table-2", ResourceType: common.Dynamodb_table, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(got, expected); diff != "" {
		t.Errorf("mismatch (-got +want):\n%s", diff)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListTables, got %d", callCount)
	}
}

func TestGetDynamodbTables_ErrorIsSkipped(t *testing.T) {
	mock := &mockDynamodbClient{
		listTablesFn: func(_ context.Context, _ *dynamodb.ListTablesInput, _ ...func(*dynamodb.Options)) (*dynamodb.ListTablesOutput, error) {
			return nil, fmt.Errorf("access denied")
		},
	}
	svc := getMockedDynamodbService(mock)
	got := svc.GetTables()
	if len(got) != 0 {
		t.Errorf("expected empty result on error, got %v", got)
	}
}
