package main

import (
	"encoding/json"
	"fmt"

	claws "github.com/noclickops/aws"
	"github.com/noclickops/common"
)

func main() {
	opts := parseFlags()

	configs := generatePerRegionConfigs(opts.regionsList)

	println("Downloading statefiles from s3")
	s3_cfg := generateStatefileBucketConfig(opts.s3BucketRegion)
	downloaded_files := download_statefiles_from_s3(opts.s3Bucket, opts.forceDownload, s3_cfg)
	if opts.stateFile != "" {
		downloaded_files = append(downloaded_files, opts.stateFile)
	}

	println("Scanning statefiles for terraform ids")
	managedIDs := getManagedIDs(downloaded_files)
	if opts.removeDownloadedStatefiles {
		defer delete_statefiles_dir()
	}

	foundRecords := make(map[string][]common.Resource)
	iamclient := claws.NewIAMClientFromConfigs(configs)
	println("Retrieving IAM policies")
	foundRecords["iam_policies"] = iamclient.GetAllPoliciesArns()
	println("Retrieving IAM users")
	foundRecords["iam_users"] = iamclient.GetAllIAMUsers()
	println("Retrieving IAM groups")
	foundRecords["iam_groups"] = iamclient.GetAllIAMGroups()

	ssmclient := claws.NewSSMClientFromConfigs(configs)
	println("Retrieving SSM params")
	foundRecords["ssm_params"] = ssmclient.GetAllParametersNames()

	route53client := claws.NewRoute53ClientFromConfigs(configs)
	println("Retrieving route53 records")
	foundRecords["route53_records"] = route53client.GetAllRoute53RecordIds()

	ec2client := claws.NewEC2ClientFromConfigs(configs)
	println("Retrieving security groups")
	foundRecords["ec2_security_groups"] = ec2client.GetAllSecurityGroups()
	println("Retrieving security group rules")
	foundRecords["ec2_security_group_rules"] = ec2client.GetAllSecurityGroupRules()

	ssoadminclient := claws.NewSSOAdminClientFromConfigs(configs)
	identitystoreclient := claws.NewIdentityStoreClientFromConfigs(configs)
	println("Retrieving identity store users")
	foundRecords["identity_store_users"] = identitystoreclient.GetAllIdentityStoreUsers(&ssoadminclient)
	println("Retrieving identity store groups")
	foundRecords["identity_store_groups"] = identitystoreclient.GetAllIdentityStoreGroups(&ssoadminclient)
	println("Retrieving permission sets")
	foundRecords["ssoadmin_permission_sets"] = ssoadminclient.GetAllPermissionSets()

	eksclient := claws.NewEKSClientFromConfigs(configs)
	println("Retrieving EKS clusters")
	foundRecords["eks_clusters"] = eksclient.GetAllEKSClusters()

	unmanagedResourceIds := filter(managedIDs, foundRecords)
	json, _ := json.Marshal(unmanagedResourceIds)
	fmt.Println(string(json))
}
