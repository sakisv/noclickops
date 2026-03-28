package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/config"
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

func filter(managedIds map[string]struct{}, listsOfFoundIds ...[]string) []string {
	var notFound []string
	for _, list := range listsOfFoundIds {
		for _, el := range list {
			_, found := managedIds[el]
			if found {
				println("Found " + el)
			} else {
				notFound = append(notFound, el)
				println("Not found " + el)
			}
		}
	}
	return notFound
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

	policies := aws.GetAllPoliciesArns(cfg)
	parameters := aws.GetAllParametersNames(cfg)
	route53_records := aws.GetAllRoute53RecordIds(cfg)

	print(filter(managedIds, policies, parameters, route53_records))
}
