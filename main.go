package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/eks"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/identitystore"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	"github.com/aws/aws-sdk-go-v2/service/ssoadmin"
	claws "github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func main() {
	var stateFile string
	var regions string
	var s3_bucket string
	var s3_bucket_region string
	flag.StringVar(&stateFile, "statefile", "", "The statefile to parse")
	flag.StringVar(&s3_bucket, "s3_bucket", "", "Download statefile(s) from this s3 bucket")
	flag.StringVar(&s3_bucket_region, "s3_bucket_region", "", "The bucket's region")
	flag.StringVar(&regions, "regions", "eu-west-1,eu-west-2", "Comma-separated list of regions to check")
	flag.Parse()

	if stateFile == "" && s3_bucket == "" {
		fmt.Println("At least one of s3_bucket or statefile must be provided")
		fmt.Println("Use -h / --help")
		return
	}

	configs := generatePerRegionConfigs(regions)

	println("Downloading statefiles from s3")
	s3_cfg := configs[0]
	if _, ok := VALID_REGIONS[strings.ToLower(strings.TrimSpace(s3_bucket_region))]; ok {
		s3_cfg.Region = strings.ToLower(strings.TrimSpace(s3_bucket_region))
	}
	downloaded_files := download_statefiles_from_s3(s3_bucket, s3_cfg)
	if stateFile != "" {
		downloaded_files = append(downloaded_files, stateFile)
	}
	println("Scanning statefiles for terraform ids")
	managedIDs := getManagedIDs(downloaded_files)
	defer delete_statefiles_dir()

	foundRecords := make(map[string][]common.Resource)
	println("Retrieving IAM policies")
	foundRecords["iam_policies"] = claws.GetAllPoliciesArns(iam.NewFromConfig(configs[0]))
	println("Retrieving IAM users")
	foundRecords["iam_users"] = claws.GetAllIAMUsers(iam.NewFromConfig(configs[0]))
	println("Retrieving IAM groups")
	foundRecords["iam_groups"] = claws.GetAllIAMGroups(iam.NewFromConfig(configs[0]))
	println("Retrieving SSM params")
	foundRecords["ssm_params"] = claws.GetAllParametersNames(ssm.NewFromConfig(configs[0]))
	println("Retrieving route53 records")
	foundRecords["route53_records"] = claws.GetAllRoute53RecordIds(route53.NewFromConfig(configs[0]))
	println("Retrieving security groups")
	foundRecords["ec2_security_groups"] = claws.GetAllSecurityGroups(ec2.NewFromConfig(configs[0]))
	println("Retrieving security group rules")
	foundRecords["ec2_security_group_rules"] = claws.GetAllSecurityGroupRules(ec2.NewFromConfig(configs[0]))
	println("Retrieving identity store users")
	foundRecords["identity_store_users"] = claws.GetAllIdentityStoreUsers(identitystore.NewFromConfig(configs[0]), ssoadmin.NewFromConfig(configs[0]))
	println("Retrieving identity store groups")
	foundRecords["identity_store_groups"] = claws.GetAllIdentityStoreGroups(identitystore.NewFromConfig(configs[0]), ssoadmin.NewFromConfig(configs[0]))
	println("Retrieving permission sets")
	foundRecords["ssoadmin_permission_sets"] = claws.GetAllPermissionSets(ssoadmin.NewFromConfig(configs[0]))
	println("Retrieving EKS clusters")
	foundRecords["eks_clusters"] = claws.GetAllEKSClusters(eks.NewFromConfig(configs[0]))

	unmanagedResourceIds := filter(managedIDs, foundRecords)
	json, _ := json.Marshal(unmanagedResourceIds)
	fmt.Println(string(json))
}
