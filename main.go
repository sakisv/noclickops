package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"os"
	"regexp"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/iam/types"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func main() {
	var stateFile string
	var region string
	flag.StringVar(&stateFile, "statefile", "", "The statefile to parse")
	flag.StringVar(&region, "region", "eu-west-1", "The AWS region to target")
	flag.Parse()
	reg := regexp.MustCompile(`\"id\": \".*\",?`)

	contents_b, err := os.ReadFile(stateFile)
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
	fmt.Println(managed_ids)

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	ssm_client := ssm.NewFromConfig(cfg)

	res, err := ssm_client.GetParametersByPath(context.TODO(), &ssm.GetParametersByPathInput{
		Path:       aws.String("/"),
		MaxResults: aws.Int32(10),
		Recursive:  aws.Bool(true),
	})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(len(res.Parameters))
	fmt.Println("Found " + strconv.Itoa(len(res.Parameters)) + " parameters")
	for _, el := range res.Parameters {
		fmt.Println(*el.Name)
	}

	iam_client := iam.NewFromConfig(cfg)

	res_iam, err := iam_client.ListPolicies(context.TODO(), &iam.ListPoliciesInput{
		MaxItems: aws.Int32(500),
		Scope:    types.PolicyScopeTypeLocal,
	})

	if err != nil {
		log.Fatal(err)
	}
	fmt.Println("Found " + strconv.Itoa(len(res_iam.Policies)) + " policies")
	for _, el := range res_iam.Policies {
		_, ok := managed_ids[*el.Arn]
		if !ok {
			println("Unmanaged policy: " + *el.Arn)
		} else {
			println("Managed policy: " + *el.Arn)
		}
	}
}
