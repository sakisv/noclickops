package aws

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/route53/types"
)

func GetAllRoute53RecordIds(cfg aws.Config) []string {
	client := route53.NewFromConfig(cfg)
	hostedZones, err := client.ListHostedZones(context.TODO(), &route53.ListHostedZonesInput{})
	if err != nil {
		log.Fatal(err)
	}

	var ids []string
	var id string
	for _, zone := range hostedZones.HostedZones {
		if *zone.ResourceRecordSetCount == 0 {
			continue
		}

		zone_id := strings.Split(*zone.Id, "/")[2]
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

				ids = append(ids, id)
			}

			if !listRecordSetsResponse.IsTruncated {
				break
			}
			nextRecordIdentifier = listRecordSetsResponse.NextRecordIdentifier
			nextRecordName = listRecordSetsResponse.NextRecordName
			nextRecordType = listRecordSetsResponse.NextRecordType
		}
	}

	return ids
}
