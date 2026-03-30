package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/google/go-cmp/cmp"
	"github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func TestListPolicies_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockIAMClient{
		listPoliciesFn: func(_ context.Context, params *iam.ListPoliciesInput, _ ...func(*iam.Options)) (*iam.ListPoliciesOutput, error) {
			callCount++
			if callCount == 1 {
				return &iam.ListPoliciesOutput{
					IsTruncated: true,
					Marker:      ptr("next"),
					Policies: []types.Policy{
						{Arn: ptr("arn:policy_1")},
					},
				}, nil
			}
			// assert pagination params were passed through correctly
			if params.Marker == nil || *params.Marker != "next" {
				return nil, fmt.Errorf("Wrong Marker, expected 'next' got '%v'", *params.Marker)
			}
			return &iam.ListPoliciesOutput{
				IsTruncated: false,
				Policies: []types.Policy{
					{Arn: ptr("arn:policy_2")},
				},
			}, nil
		},
	}
	ids := aws.GetAllPoliciesArns(mock)
	expected := []common.Resource{
		{TerraformID: "arn:policy_1", ResourceType: "iam.policy"},
		{TerraformID: "arn:policy_2", ResourceType: "iam.policy"},
	}
	if diff := cmp.Diff(ids, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, ids)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListPolicies, got %d", callCount)
	}
}
