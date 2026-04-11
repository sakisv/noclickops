package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssm/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func getMockedSSMService(mock *mockSSMClient) aws.NoclickopsSSMService {
	return aws.NoclickopsSSMService{
		Clients: []aws.NoclickopsSSMClient{
			{
				Client:     mock,
				ClientMeta: aws.ClientMeta{Region: "eu-west-1"},
			},
		},
		ServiceMeta: common.ServiceMeta{Global: false, ServiceName: "ssm"},
	}
}

func TestGetAllParametersNames(t *testing.T) {
	callCount := 0
	mock := &mockSSMClient{
		getParametersByPathFn: func(ctx context.Context, params *ssm.GetParametersByPathInput, optFns ...func(*ssm.Options)) (*ssm.GetParametersByPathOutput, error) {
			callCount++
			if callCount == 1 {
				return &ssm.GetParametersByPathOutput{
					NextToken: ptr("next"),
					Parameters: []types.Parameter{
						{ARN: ptr("some:arn:"), DataType: ptr("string"), Name: ptr("/some/parameter")},
					},
				}, nil
			}

			if *params.NextToken != "next" {
				return nil, fmt.Errorf("Wrong NextToken. Expected 'next' got '%v'", *params.NextToken)
			}

			return &ssm.GetParametersByPathOutput{
				Parameters: []types.Parameter{
					{ARN: ptr("second:arn:"), DataType: ptr("string"), Name: ptr("/some/other/parameter")},
				},
			}, nil

		},
	}
	client := getMockedSSMService(mock)
	ids := client.GetAllParametersNames()
	expected := []common.Resource{
		{TerraformID: "/some/parameter", ResourceType: common.SSM_parameter, Region: "eu-west-1"},
		{TerraformID: "/some/other/parameter", ResourceType: common.SSM_parameter, Region: "eu-west-1"},
	}
	if diff := cmp.Diff(ids, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, ids)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to GetAllParametersNames, got %d", callCount)
	}
}
