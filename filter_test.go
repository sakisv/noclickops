package main

import (
	"context"
	"slices"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi"
	rgtypes "github.com/aws/aws-sdk-go-v2/service/resourcegroupstaggingapi/types"
	claws "github.com/noclickops/aws"
)

type mockRGTaggingClient struct {
	getTagKeysFn   func(context.Context, *resourcegroupstaggingapi.GetTagKeysInput, ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetTagKeysOutput, error)
	getResourcesFn func(context.Context, *resourcegroupstaggingapi.GetResourcesInput, ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error)
}

func (m *mockRGTaggingClient) GetTagKeys(ctx context.Context, params *resourcegroupstaggingapi.GetTagKeysInput, optFns ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetTagKeysOutput, error) {
	return m.getTagKeysFn(ctx, params, optFns...)
}

func (m *mockRGTaggingClient) GetResources(ctx context.Context, params *resourcegroupstaggingapi.GetResourcesInput, optFns ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
	return m.getResourcesFn(ctx, params, optFns...)
}

func newRGTaggingService(tagKeys []string, resourcesByTagKey map[string][]string) claws.NoclickopsResourceGroupTaggingAPIService {
	mock := &mockRGTaggingClient{
		getTagKeysFn: func(_ context.Context, _ *resourcegroupstaggingapi.GetTagKeysInput, _ ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetTagKeysOutput, error) {
			return &resourcegroupstaggingapi.GetTagKeysOutput{TagKeys: tagKeys}, nil
		},
		getResourcesFn: func(_ context.Context, params *resourcegroupstaggingapi.GetResourcesInput, _ ...func(*resourcegroupstaggingapi.Options)) (*resourcegroupstaggingapi.GetResourcesOutput, error) {
			var mappings []rgtypes.ResourceTagMapping
			if len(params.TagFilters) > 0 {
				key := *params.TagFilters[0].Key
				for _, arn := range resourcesByTagKey[key] {
					arnCopy := arn
					mappings = append(mappings, rgtypes.ResourceTagMapping{ResourceARN: &arnCopy})
				}
			}
			return &resourcegroupstaggingapi.GetResourcesOutput{ResourceTagMappingList: mappings}, nil
		},
	}
	return claws.NoclickopsResourceGroupTaggingAPIService{
		Clients: []claws.NoclickopsResourceGroupTaggingAPIClient{
			{Client: mock},
		},
	}
}

func TestGetIgnoredTagResources(t *testing.T) {
	origPrefixes := IGNORED_TAG_KEY_PREFIXES
	defer func() { IGNORED_TAG_KEY_PREFIXES = origPrefixes }()
	IGNORED_TAG_KEY_PREFIXES = []string{"test-prefix/"}

	tests := []struct {
		name              string
		tagKeys           []string
		resourcesByTagKey map[string][]string
		ignoredTags       map[string][]string
		wantArns          map[string]struct{}
		// wantIgnoredTags, if set, is checked against ignoredTags after the call to
		// verify the map was mutated as expected.
		wantIgnoredTags map[string][]string
	}{
		{
			name:              "returns empty map when no tags and no defaults",
			tagKeys:           []string{},
			resourcesByTagKey: map[string][]string{},
			ignoredTags:       map[string][]string{},
			wantArns:          map[string]struct{}{},
		},
		{
			name:    "returns ARNs for explicitly provided tag",
			tagKeys: []string{},
			resourcesByTagKey: map[string][]string{
				"my-tag": {"arn:aws:ec2:us-east-1:123456789:instance/i-abc123"},
			},
			ignoredTags: map[string][]string{"my-tag": {"value"}},
			wantArns: map[string]struct{}{
				"arn:aws:ec2:us-east-1:123456789:instance/i-abc123": {},
			},
		},
		{
			name:              "returns empty when explicit tag matches no resources",
			tagKeys:           []string{},
			resourcesByTagKey: map[string][]string{},
			ignoredTags:       map[string][]string{"my-tag": {"value"}},
			wantArns:          map[string]struct{}{},
		},
		{
			name:    "default tag keys are added and their resources collected",
			tagKeys: []string{"test-prefix/cluster-a", "unrelated-key"},
			resourcesByTagKey: map[string][]string{
				"test-prefix/cluster-a": {"arn:aws:ec2:us-east-1:123:vpc/vpc-1"},
			},
			ignoredTags: map[string][]string{},
			wantArns: map[string]struct{}{
				"arn:aws:ec2:us-east-1:123:vpc/vpc-1": {},
			},
			wantIgnoredTags: map[string][]string{
				"test-prefix/cluster-a": {},
			},
		},
		{
			name:    "existing values for a default tag key are not overridden",
			tagKeys: []string{"test-prefix/cluster"},
			resourcesByTagKey: map[string][]string{
				"test-prefix/cluster": {"arn:aws:ec2:us-east-1:123:subnet/subnet-1"},
			},
			ignoredTags: map[string][]string{
				"test-prefix/cluster": {"owned"},
			},
			wantArns: map[string]struct{}{
				"arn:aws:ec2:us-east-1:123:subnet/subnet-1": {},
			},
			wantIgnoredTags: map[string][]string{
				"test-prefix/cluster": {"owned"},
			},
		},
		{
			name:    "deduplicates ARNs that appear under multiple tags",
			tagKeys: []string{},
			resourcesByTagKey: map[string][]string{
				"tag-a": {
					"arn:aws:ec2:us-east-1:123:instance/i-shared",
					"arn:aws:ec2:us-east-1:123:instance/i-unique",
				},
				"tag-b": {"arn:aws:ec2:us-east-1:123:instance/i-shared"},
			},
			ignoredTags: map[string][]string{
				"tag-a": {},
				"tag-b": {},
			},
			wantArns: map[string]struct{}{
				"arn:aws:ec2:us-east-1:123:instance/i-shared": {},
				"arn:aws:ec2:us-east-1:123:instance/i-unique": {},
			},
		},
		{
			name:    "collects ARNs from both explicit and default tags",
			tagKeys: []string{"test-prefix/cluster"},
			resourcesByTagKey: map[string][]string{
				"test-prefix/cluster": {"arn:aws:ec2:us-east-1:123:vpc/vpc-default"},
				"explicit-tag":        {"arn:aws:ec2:us-east-1:123:vpc/vpc-explicit"},
			},
			ignoredTags: map[string][]string{
				"explicit-tag": {"true"},
			},
			wantArns: map[string]struct{}{
				"arn:aws:ec2:us-east-1:123:vpc/vpc-default":  {},
				"arn:aws:ec2:us-east-1:123:vpc/vpc-explicit": {},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc := newRGTaggingService(tt.tagKeys, tt.resourcesByTagKey)
			got := getIgnoredTagResources(svc, tt.ignoredTags)

			if len(got) != len(tt.wantArns) {
				t.Errorf("got %d ARNs, want %d\ngot:  %v\nwant: %v", len(got), len(tt.wantArns), got, tt.wantArns)
				return
			}
			for arn := range tt.wantArns {
				if _, found := got[arn]; !found {
					t.Errorf("missing ARN %q in result %v", arn, got)
				}
			}

			for k, wantVals := range tt.wantIgnoredTags {
				gotVals, found := tt.ignoredTags[k]
				if !found {
					t.Errorf("expected key %q in ignoredTags after call", k)
					continue
				}
				if !slices.Equal(gotVals, wantVals) {
					t.Errorf("ignoredTags[%q] = %v, want %v", k, gotVals, wantVals)
				}
			}
		})
	}
}
