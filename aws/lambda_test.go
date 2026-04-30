package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/lambda"
	ltypes "github.com/aws/aws-sdk-go-v2/service/lambda/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedLambdaService(clients ...aws.NoclickopsLambdaClient) aws.NoclickopsLambdaService {
	return aws.NoclickopsLambdaService{
		Clients:     clients,
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "lambda"},
	}
}

func TestGetAllLambdaFunctions(t *testing.T) {
	t.Run("returns all functions from a single page", func(t *testing.T) {
		mock := &mockLambdaClient{
			listFunctionsFn: func(_ context.Context, _ *lambda.ListFunctionsInput, _ ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
				return &lambda.ListFunctionsOutput{
					Functions: []ltypes.FunctionConfiguration{
						{FunctionArn: ptr("arn:aws:lambda:eu-west-1:123:function:my-func"), FunctionName: ptr("my-func")},
						{FunctionArn: ptr("arn:aws:lambda:eu-west-1:123:function:other-func"), FunctionName: ptr("other-func")},
					},
				}, nil
			},
		}
		svc := getMockedLambdaService(aws.NoclickopsLambdaClient{Client: mock, ClientMeta: aws.ClientMeta{Region: "eu-west-1"}})
		got := svc.GetAllLambdaFunctions()
		want := []common.Resource{
			{Arn: "arn:aws:lambda:eu-west-1:123:function:my-func", TerraformID: "my-func", ResourceType: common.Lambda_function, Region: "eu-west-1"},
			{Arn: "arn:aws:lambda:eu-west-1:123:function:other-func", TerraformID: "other-func", ResourceType: common.Lambda_function, Region: "eu-west-1"},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("unexpected result (-got +want):\n%s", diff)
		}
	})

	t.Run("follows pagination via NextMarker", func(t *testing.T) {
		callCount := 0
		mock := &mockLambdaClient{
			listFunctionsFn: func(_ context.Context, params *lambda.ListFunctionsInput, _ ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
				callCount++
				if callCount == 1 {
					return &lambda.ListFunctionsOutput{
						NextMarker: ptr("page2"),
						Functions: []ltypes.FunctionConfiguration{
							{FunctionArn: ptr("arn:aws:lambda:eu-west-1:123:function:func-1"), FunctionName: ptr("func-1")},
						},
					}, nil
				}
				if params.Marker == nil || *params.Marker != "page2" {
					return nil, fmt.Errorf("expected Marker 'page2', got %v", params.Marker)
				}
				return &lambda.ListFunctionsOutput{
					Functions: []ltypes.FunctionConfiguration{
						{FunctionArn: ptr("arn:aws:lambda:eu-west-1:123:function:func-2"), FunctionName: ptr("func-2")},
					},
				}, nil
			},
		}
		svc := getMockedLambdaService(aws.NoclickopsLambdaClient{Client: mock, ClientMeta: aws.ClientMeta{Region: "eu-west-1"}})
		got := svc.GetAllLambdaFunctions()
		want := []common.Resource{
			{Arn: "arn:aws:lambda:eu-west-1:123:function:func-1", TerraformID: "func-1", ResourceType: common.Lambda_function, Region: "eu-west-1"},
			{Arn: "arn:aws:lambda:eu-west-1:123:function:func-2", TerraformID: "func-2", ResourceType: common.Lambda_function, Region: "eu-west-1"},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("unexpected result (-got +want):\n%s", diff)
		}
		if callCount != 2 {
			t.Errorf("expected 2 calls to ListFunctions, got %d", callCount)
		}
	})

	t.Run("collects functions across multiple regions", func(t *testing.T) {
		makeClient := func(region, arn, name string) aws.NoclickopsLambdaClient {
			return aws.NoclickopsLambdaClient{
				ClientMeta: aws.ClientMeta{Region: region},
				Client: &mockLambdaClient{
					listFunctionsFn: func(_ context.Context, _ *lambda.ListFunctionsInput, _ ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
						return &lambda.ListFunctionsOutput{
							Functions: []ltypes.FunctionConfiguration{
								{FunctionArn: ptr(arn), FunctionName: ptr(name)},
							},
						}, nil
					},
				},
			}
		}
		svc := getMockedLambdaService(
			makeClient("eu-west-1", "arn:aws:lambda:eu-west-1:123:function:west-func", "west-func"),
			makeClient("us-east-1", "arn:aws:lambda:us-east-1:123:function:east-func", "east-func"),
		)
		got := svc.GetAllLambdaFunctions()
		want := []common.Resource{
			{Arn: "arn:aws:lambda:eu-west-1:123:function:west-func", TerraformID: "west-func", ResourceType: common.Lambda_function, Region: "eu-west-1"},
			{Arn: "arn:aws:lambda:us-east-1:123:function:east-func", TerraformID: "east-func", ResourceType: common.Lambda_function, Region: "us-east-1"},
		}
		if diff := cmp.Diff(got, want); diff != "" {
			t.Errorf("unexpected result (-got +want):\n%s", diff)
		}
	})

	t.Run("returns empty slice when no functions exist", func(t *testing.T) {
		mock := &mockLambdaClient{
			listFunctionsFn: func(_ context.Context, _ *lambda.ListFunctionsInput, _ ...func(*lambda.Options)) (*lambda.ListFunctionsOutput, error) {
				return &lambda.ListFunctionsOutput{}, nil
			},
		}
		svc := getMockedLambdaService(aws.NoclickopsLambdaClient{Client: mock, ClientMeta: aws.ClientMeta{Region: "eu-west-1"}})
		got := svc.GetAllLambdaFunctions()
		if len(got) != 0 {
			t.Errorf("expected empty result, got %v", got)
		}
	})
}
