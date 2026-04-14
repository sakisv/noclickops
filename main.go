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

	var downloaded_files []string
	if opts.s3Bucket != "" {
		println("Downloading statefiles from s3")
		s3_cfg := generateStatefileBucketConfig(opts.s3BucketRegion)
		downloaded_files = download_statefiles_from_s3(opts.s3Bucket, opts.forceDownload, s3_cfg)

	}
	if opts.stateFile != "" {
		downloaded_files = append(downloaded_files, opts.stateFile)
	}

	println("Scanning statefiles for terraform ids")
	managedIDs := getManagedIDs(downloaded_files)
	if opts.removeDownloadedStatefiles {
		defer delete_statefiles_dir()
	}

	iamclient := claws.NewNoclickopsServiceFromConfigs(common.IAM, configs)
	foundResources := make(map[string][]common.Resource)
	foundResources[iamclient.GetServiceName()] = iamclient.GetAllResources()

	ssmclient := claws.NewNoclickopsServiceFromConfigs(common.SSM, configs)
	foundResources[ssmclient.GetServiceName()] = ssmclient.GetAllResources()

	route53client := claws.NewNoclickopsServiceFromConfigs(common.Route53, configs)
	foundResources[route53client.GetServiceName()] = route53client.GetAllResources()

	ec2client := claws.NewNoclickopsServiceFromConfigs(common.EC2, configs)
	foundResources[ec2client.GetServiceName()] = ec2client.GetAllResources()

	ssoadminclient := claws.NewNoclickopsServiceFromConfigs(common.SSOAdmin, configs)
	identitystoreclient := claws.NewNoclickopsServiceFromConfigs(common.IdentityStore, configs)
	foundResources[identitystoreclient.GetServiceName()] = identitystoreclient.GetAllResources()
	foundResources[ssoadminclient.GetServiceName()] = ssoadminclient.GetAllResources()

	eksclient := claws.NewNoclickopsServiceFromConfigs(common.EKS, configs)
	foundResources[eksclient.GetServiceName()] = eksclient.GetAllResources()

	rdsclient := claws.NewNoclickopsServiceFromConfigs(common.RDS, configs)
	foundResources[rdsclient.GetServiceName()] = rdsclient.GetAllResources()

	snsclient := claws.NewNoclickopsServiceFromConfigs(common.SNS, configs)
	foundResources[snsclient.GetServiceName()] = snsclient.GetAllResources()

	unmanagedResourceIds := filter(managedIDs, foundResources)
	json, _ := json.Marshal(unmanagedResourceIds)
	fmt.Println(string(json))
}
