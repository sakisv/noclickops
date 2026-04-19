package main

import (
	"context"
	"log"
	"os"
	"path"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/account"
)

var VALID_REGIONS = map[string]string{
	"af-south-1":     "Africa (Cape Town)",
	"ap-east-1":      "Asia Pacific (Hong Kong)",
	"ap-east-2":      "Asia Pacific (Taipei)",
	"ap-northeast-1": "Asia Pacific (Tokyo)",
	"ap-northeast-2": "Asia Pacific (Seoul)",
	"ap-northeast-3": "Asia Pacific (Osaka)",
	"ap-south-1":     "Asia Pacific (Mumbai)",
	"ap-south-2":     "Asia Pacific (Hyderabad)",
	"ap-southeast-1": "Asia Pacific (Singapore)",
	"ap-southeast-2": "Asia Pacific (Sydney)",
	"ap-southeast-3": "Asia Pacific (Jakarta)",
	"ap-southeast-4": "Asia Pacific (Melbourne)",
	"ap-southeast-5": "Asia Pacific (Malaysia)",
	"ap-southeast-6": "Asia Pacific (New Zealand)",
	"ap-southeast-7": "Asia Pacific (Thailand)",
	"ca-central-1":   "Canada (Central)",
	"ca-west-1":      "Canada West (Calgary)",
	"cn-north-1":     "China (Beijing)",
	"cn-northwest-1": "China (Ningxia)",
	"eu-central-1":   "Europe (Frankfurt)",
	"eu-central-2":   "Europe (Zurich)",
	"eu-north-1":     "Europe (Stockholm)",
	"eu-south-1":     "Europe (Milan)",
	"eu-south-2":     "Europe (Spain)",
	"eu-west-1":      "Europe (Ireland)",
	"eu-west-2":      "Europe (London)",
	"eu-west-3":      "Europe (Paris)",
	"il-central-1":   "Israel (Tel Aviv)",
	"me-central-1":   "Middle East (UAE)",
	"me-south-1":     "Middle East (Bahrain)",
	"mx-central-1":   "Mexico (Central)",
	"sa-east-1":      "South America (Sao Paolo)",
	"us-east-1":      "US East (N. Virginia)",
	"us-east-2":      "US East (Ohio)",
	"us-west-1":      "US West (N. California)",
	"us-west-2":      "US West (Oregon)",
	//  US gov cloud
	"us-gov-east-1": "AWS GovCloud (US-East)",
	"us-gov-west-1": "AWS GovCloud (US-West)",
	//  european sovereign cloud
	"eusc-de-east-1": "AWS European Sovereign Cloud (Germany)",
}

func getAllRegions() []string {
	keys := make([]string, 0, len(VALID_REGIONS))
	for k := range VALID_REGIONS {
		keys = append(keys, k)
	}

	return keys
}

func getFullPathToHomeTarget(to string) string {
	homedir, _ := os.UserHomeDir()
	return path.Join(homedir, to)
}

func getConfigForRegion(region string) aws.Config {
	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}
	cfg.Region = region
	return cfg
}

func generatePerRegionConfigs(regions []string) []aws.Config {
	var configs []aws.Config
	for _, r := range regions {
		cfg, err := config.LoadDefaultConfig(context.TODO())
		if err != nil {
			log.Fatal(err)
		}

		cfg.Region = r
		configs = append(configs, cfg)
	}

	return configs
}

func getAccountId() string {
	cfg := getConfigForRegion("us-east-1")
	acc := account.NewFromConfig(cfg)
	res, err := acc.GetAccountInformation(context.TODO(), &account.GetAccountInformationInput{})

	if err != nil {
		log.Fatal(err)
	}

	return *res.AccountId
}
