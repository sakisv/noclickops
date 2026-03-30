package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
	"github.com/noclickops/common"
)

type Route53Client interface {
	ListHostedZones(ctx context.Context, params *route53.ListHostedZonesInput, optFns ...func(*route53.Options)) (*route53.ListHostedZonesOutput, error)
	ListResourceRecordSets(ctx context.Context, params *route53.ListResourceRecordSetsInput, optFns ...func(*route53.Options)) (*route53.ListResourceRecordSetsOutput, error)
}

func GetAllRoute53RecordIds(client Route53Client) []common.Resource {
	hostedZones, err := client.ListHostedZones(context.TODO(), &route53.ListHostedZonesInput{})
	if err != nil {
		log.Fatal(err)
	}

	var resources []common.Resource
	var id string
	for _, zone := range hostedZones.HostedZones {
		zone_id := strings.Split(*zone.Id, "/")[2]
		resources = append(resources, common.Resource{TerraformID: zone_id, ResourceType: common.Route53_zone})
		if *zone.ResourceRecordSetCount == 0 {
			continue
		}

		var nextRecordName *string
		var nextRecordIdentifier *string
		var nextRecordType types.RRType
		for {
			listRecordSetsResponse, err := client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
				HostedZoneId:          zone.Id,
				StartRecordIdentifier: nextRecordIdentifier,
				StartRecordName:       nextRecordName,
				StartRecordType:       nextRecordType,
			})

			if err != nil {
				log.Fatal(err)
			}

			for _, record := range listRecordSetsResponse.ResourceRecordSets {
				record_name := strings.TrimSuffix(*record.Name, *zone.Name)
				record_name = strings.TrimSuffix(record_name, ".")
				if record.SetIdentifier != nil && *record.SetIdentifier != "" {
					id = fmt.Sprintf("%v_%v_%v_%v", zone_id, record_name, record.Type, *record.SetIdentifier)
				} else {
					id = fmt.Sprintf("%v_%v_%v", zone_id, record_name, record.Type)
				}

				resources = append(resources, common.Resource{TerraformID: id, ResourceType: common.Route53_record})
			}

			if !listRecordSetsResponse.IsTruncated {
				break
			}
			nextRecordIdentifier = listRecordSetsResponse.NextRecordIdentifier
			nextRecordName = listRecordSetsResponse.NextRecordName
			nextRecordType = listRecordSetsResponse.NextRecordType
		}
	}

	return resources
}
