package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
)

type options struct {
	stateFile                  string
	removeDownloadedStatefiles bool
	forceDownload              bool
	regions                    string
	s3Bucket                   string
	s3BucketRegion             string
	regionsList                []string
}

func (opts *options) validate() error {
	var errs []string

	if opts.stateFile == "" && opts.s3Bucket == "" {
		errs = append(errs, "At least one of 's3-bucket' or 'statefile' must be provided")
	}

	if opts.s3Bucket != "" {
		if opts.s3BucketRegion == "" {
			errs = append(errs, "s3-bucket-region must be provided if s3-bucket is defined")
		}
		if opts.s3BucketRegion == "all" {
			errs = append(errs, "s3-bucket-region cannot be 'all'")
		}
		if !isValidRegion(opts.s3BucketRegion) {
			errs = append(errs, fmt.Sprintf("'%v' is not a valid region", opts.s3BucketRegion))
		}
	}

	if opts.forceDownload && opts.s3Bucket == "" {
		errs = append(errs, "'force-download' must be used alongside an 's3-bucket'")
	}

	regions := strings.Split(opts.regions, ",")
	for _, region := range regions {
		r := strings.ToLower(strings.TrimSpace(region))
		if !isValidRegion(r) {
			errs = append(errs, fmt.Sprintf("'%v' is not a valid region", r))
			continue
		}

		// if any of the regions is "all" then we overwrite what we currently have and break
		if r == "all" {
			opts.regionsList = getAllRegions()
			break
		} else {
			opts.regionsList = append(opts.regionsList, r)
		}
	}

	if len(errs) > 0 {
		errs = append(errs, "Use -h / --help")
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func isValidRegion(region string) bool {
	if region == "all" {
		return true
	}
	_, found := VALID_REGIONS[region]
	return found
}

func parseFlags() options {
	var opts options

	flag.StringVar(&opts.stateFile, "statefile", "", "The statefile to parse")
	flag.StringVar(&opts.s3Bucket, "s3-bucket", "", "Download statefile(s) from this s3 bucket")
	flag.StringVar(&opts.s3BucketRegion, "s3-bucket-region", "", "The bucket's region. Cannot be 'all'")
	flag.StringVar(&opts.regions, "regions", "all", "Comma-separated list of regions to check, or 'all'")
	flag.BoolVar(&opts.removeDownloadedStatefiles, "remove-downloaded-statefiles", false, "If specified, any downloaded statefiles will be deleted at the end")
	flag.BoolVar(&opts.forceDownload, "force-download", false, "If specified, it will download all the files from the bucket even they overwrite existing ones")
	flag.Parse()

	err := opts.validate()
	if err != nil {
		log.Fatal(err)
	}

	return opts
}
