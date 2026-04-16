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

	foundResources := make(map[string][]common.Resource)
	for service := range claws.SERVICES {
		if service == common.ResourceGroupsTaggingAPI {
			continue
		}
		client := claws.NewNoclickopsServiceFromConfigs(service, configs)
		println("Fetching resources for " + client.GetServiceName())
		foundResources[client.GetServiceName()] = client.GetAllResources()
	}

	var ignoredArns []string
	if len(opts.ignoreTagsMap) > 0 {
		ignoredArns = getIgnoredTagResources(opts.ignoreTagsMap, configs)
	}
	unmanagedResourceIds := filter(managedIDs, foundResources, ignoredArns)
	json, _ := json.Marshal(unmanagedResourceIds)
	fmt.Println(string(json))
}
