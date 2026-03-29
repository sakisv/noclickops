package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/noclickops/aws"
)

func getManagedIds(statefile_path string) map[string]struct{} {
	reg := regexp.MustCompile(`\"id\": \".*\",?`)

	contents_b, err := os.ReadFile(statefile_path)
	if err != nil {
		log.Fatal(err)
	}
	contents := string(contents_b[:])
	finds := reg.FindAllString(contents, -1)
	fmt.Println()

	var managed_ids = make(map[string]struct{})
	for _, el := range finds {
		res := strings.Split(el, "\": ")
		if len(res) != 2 {
			continue
		}
		managed_id, _ := strings.CutSuffix(res[1], ",")
		managed_id = strings.ReplaceAll(managed_id, "\"", "")
		_, ok := managed_ids[managed_id]
		if !ok {
			managed_ids[managed_id] = struct{}{}
		}
	}
	return managed_ids
}

func filter(managedIds map[string]struct{}, foundRecords map[string][]string) map[string][]string {
	unmanagedResourceIds := make(map[string][]string)
	for key, value := range foundRecords {
		if len(value) == 0 {
			continue
		}
		for _, el := range value {
			_, found := managedIds[el]
			if found {
				println("Found " + el)
			} else {
				unmanagedResourceIds[key] = append(unmanagedResourceIds[key], el)
				println("Not found " + el)
			}
		}
	}
	return unmanagedResourceIds
}

func main() {
	var stateFile string
	var region string
	flag.StringVar(&stateFile, "statefile", "", "The statefile to parse")
	flag.StringVar(&region, "region", "eu-west-1", "The AWS region to target")
	flag.Parse()

	managedIds := getManagedIds(stateFile)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	foundRecords := make(map[string][]string)
	foundRecords["policies"] = aws.GetAllPoliciesArns(iam.NewFromConfig(cfg))
	foundRecords["ssm_params"] = aws.GetAllParametersNames(ssm.NewFromConfig(cfg))
	foundRecords["route53_records"] = aws.GetAllRoute53RecordIds(route53.NewFromConfig(cfg))

	unmanagedResourceIds := filter(managedIds, foundRecords)
	json, err := json.Marshal(unmanagedResourceIds)
	fmt.Println(string(json))
}
