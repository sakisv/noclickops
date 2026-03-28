package aws

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/route53"
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

		recordsets, err := client.ListResourceRecordSets(context.TODO(), &route53.ListResourceRecordSetsInput{
			HostedZoneId: zone.Id,
		})

		if err != nil {
			log.Fatal(err)
		}

		for _, record := range recordsets.ResourceRecordSets {
			if *record.SetIdentifier != "" {
				id = fmt.Sprintf("%v_%v_%v_%v", *zone.Id, *record.Name, record.Type, *record.SetIdentifier)
			} else {
				id = fmt.Sprintf("%v_%v_%v", *zone.Id, *record.Name, record.Type)
			}

			ids = append(ids, id)
		}
	}

	//fmt.Println("Found " + strconv.Itoa(len(ids)) + " route53 records")
	return ids
}
