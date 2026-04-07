package main

import (
	"errors"
	"flag"
	"log"
)

type options struct {
	stateFile      string
	regions        string
	s3Bucket       string
	s3BucketRegion string
}

func (opts *options) validate() error {
	err := ""

	if opts.stateFile == "" && opts.s3Bucket == "" {
		err = "At least one of 's3_bucket' or 'statefile' must be provided"
	}

	if err != "" {
		err += "\nUse -h / --help"
		return errors.New(err)
	}
	return nil
}

func parseFlags() options {
	var opts options

	flag.StringVar(&opts.stateFile, "statefile", "", "The statefile to parse")
	flag.StringVar(&opts.s3Bucket, "s3_bucket", "", "Download statefile(s) from this s3 bucket")
	flag.StringVar(&opts.s3BucketRegion, "s3_bucket_region", "", "The bucket's region")
	flag.StringVar(&opts.regions, "regions", "all", "Comma-separated list of regions to check")
	flag.Parse()

	err := opts.validate()
	if err != nil {
		log.Fatal(err)
	}

	return opts
}
