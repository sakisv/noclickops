package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"
)

type options struct {
	stateFile      string
	regions        string
	s3Bucket       string
	s3BucketRegion string
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

	regions := strings.Split(opts.regions, ",")
	for _, region := range regions {
		if !isValidRegion(region) {
			errs = append(errs, fmt.Sprintf("'%v' is not a valid region", region))
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
	flag.Parse()

	err := opts.validate()
	if err != nil {
		log.Fatal(err)
	}

	return opts
}
