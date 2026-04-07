package main

import (
	"encoding/json"
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
	opts := parseFlags()

	configs := generatePerRegionConfigs(opts.regions)

	println("Downloading statefiles from s3")
	s3_cfg := configs[0]
	if _, ok := VALID_REGIONS[strings.ToLower(strings.TrimSpace(opts.s3BucketRegion))]; ok {
		s3_cfg.Region = strings.ToLower(strings.TrimSpace(opts.s3BucketRegion))
	}
	downloaded_files := download_statefiles_from_s3(opts.s3Bucket, s3_cfg)
	if opts.stateFile != "" {
		downloaded_files = append(downloaded_files, opts.stateFile)
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
