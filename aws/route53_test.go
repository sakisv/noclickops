package aws_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/google/go-cmp/cmp"

	"github.com/noclickops/aws"
)

func TestGetAllRoute53RecordIds_NoZones(t *testing.T) {
	mock := &mockRoute53Client{
		listHostedZonesFn: func(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error) {
			return &route53.ListHostedZonesOutput{HostedZones: []types.HostedZone{}}, nil
		},
	}
	ids := aws.GetAllRoute53RecordIds(mock)
	if len(ids) != 0 {
		t.Errorf("expected empty, got %v", ids)
	}
}

func TestGetAllRoute53RecordIds_SkipsEmptyZones(t *testing.T) {
	mock := &mockRoute53Client{
		listHostedZonesFn: func(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error) {
			return &route53.ListHostedZonesOutput{
				HostedZones: []types.HostedZone{
					{Id: ptr("/hostedzone/Z123"), Name: ptr("example.com."), ResourceRecordSetCount: ptr(int64(0))},
				},
			}, nil
		},
		// listResourceRecordSetsFn intentionally nil — should never be called
	}
	ids := aws.GetAllRoute53RecordIds(mock)
	if len(ids) != 0 {
		t.Errorf("expected empty, got %v", ids)
	}
}

func TestGetAllRoute53RecordIds_SimpleRecord(t *testing.T) {
	mock := &mockRoute53Client{
		listHostedZonesFn: func(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error) {
			return &route53.ListHostedZonesOutput{
				HostedZones: []types.HostedZone{
					{Id: ptr("/hostedzone/Z123"), Name: ptr("example.com."), ResourceRecordSetCount: ptr(int64(1))},
				},
			}, nil
		},
		listResourceRecordSetsFn: func(_ context.Context, params *route53.ListResourceRecordSetsInput, _ ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
			return &route53.ListResourceRecordSetsOutput{
				IsTruncated: false,
				ResourceRecordSets: []types.ResourceRecordSet{
					{Name: ptr("www.example.com."), Type: types.RRTypeA},
				},
			}, nil
		},
	}
	ids := aws.GetAllRoute53RecordIds(mock)
	expected := []string{"Z123_www_A"}
	if diff := cmp.Diff(ids, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, ids)
	}
}

func TestGetAllRoute53RecordIds_WithSetIdentifier(t *testing.T) {
	mock := &mockRoute53Client{
		listHostedZonesFn: func(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error) {
			return &route53.ListHostedZonesOutput{
				HostedZones: []types.HostedZone{
					{Id: ptr("/hostedzone/Z123"), Name: ptr("example.com."), ResourceRecordSetCount: ptr(int64(1))},
				},
			}, nil
		},
		listResourceRecordSetsFn: func(_ context.Context, params *route53.ListResourceRecordSetsInput, _ ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
			return &route53.ListResourceRecordSetsOutput{
				IsTruncated: false,
				ResourceRecordSets: []types.ResourceRecordSet{
					{Name: ptr("www.example.com."), Type: types.RRTypeA, SetIdentifier: ptr("primary")},
				},
			}, nil
		},
	}
	ids := aws.GetAllRoute53RecordIds(mock)
	expected := []string{"Z123_www_A_primary"}
	if diff := cmp.Diff(ids, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, ids)
	}
}

func TestGetAllRoute53RecordIds_PaginationFollowed(t *testing.T) {
	callCount := 0
	mock := &mockRoute53Client{
		listHostedZonesFn: func(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error) {
			return &route53.ListHostedZonesOutput{
				HostedZones: []types.HostedZone{
					{Id: ptr("/hostedzone/Z123"), Name: ptr("example.com."), ResourceRecordSetCount: ptr(int64(2))},
				},
			}, nil
		},
		listResourceRecordSetsFn: func(_ context.Context, params *route53.ListResourceRecordSetsInput, _ ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error) {
			callCount++
			if callCount == 1 {
				return &route53.ListResourceRecordSetsOutput{
					IsTruncated:    true,
					NextRecordName: ptr("b.example.com."),
					NextRecordType: types.RRTypeA,
					ResourceRecordSets: []types.ResourceRecordSet{
						{Name: ptr("a.example.com."), Type: types.RRTypeA},
					},
				}, nil
			}
			// assert pagination params were passed through correctly
			if params.StartRecordName == nil || *params.StartRecordName != "b.example.com." {
				return nil, fmt.Errorf("expected StartRecordName=b.example.com., got %v", params.StartRecordName)
			}
			return &route53.ListResourceRecordSetsOutput{
				IsTruncated: false,
				ResourceRecordSets: []types.ResourceRecordSet{
					{Name: ptr("b.example.com."), Type: types.RRTypeA},
				},
			}, nil
		},
	}
	ids := aws.GetAllRoute53RecordIds(mock)
	expected := []string{"Z123_a_A", "Z123_b_A"}
	if diff := cmp.Diff(ids, expected); diff != "" {
		t.Errorf("expected %v, got %v", expected, ids)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls to ListResourceRecordSets, got %d", callCount)
	}
}
