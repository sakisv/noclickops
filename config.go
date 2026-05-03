package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type options struct {
	stateFile                  string
	removeDownloadedStatefiles bool
	forceDownload              bool
	regions                    string
	s3Bucket                   string
	s3BucketRegion             string
	ignoreTags                 string
	regionsList                []string
	ignoreTagsMap              map[string][]string
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
		if !isValidRegion(opts.s3BucketRegion) {
			errs = append(errs, fmt.Sprintf("'%v' is not a valid region", opts.s3BucketRegion))
		}
	}

	if opts.forceDownload && opts.s3Bucket == "" {
		errs = append(errs, "'force-download' must be used alongside an 's3-bucket'")
	}

	for region := range strings.SplitSeq(opts.regions, ",") {
		r := strings.ToLower(strings.TrimSpace(region))
		if !isValidRegion(r) {
			errs = append(errs, fmt.Sprintf("'%v' is not a valid region", r))
			continue
		}

		opts.regionsList = append(opts.regionsList, r)
	}

	opts.ignoreTagsMap = parseTags(opts.ignoreTags)

	if len(errs) > 0 {
		errs = append(errs, "Use -h / --help")
		return errors.New(strings.Join(errs, "\n"))
	}
	return nil
}

func parseTags(tags string) map[string][]string {
	parsedTags := make(map[string][]string)

	for tag := range strings.SplitSeq(tags, ",") {
		if !strings.Contains(tag, "=") {
			continue
		}

		pair := strings.Split(tag, "=")
		if len(pair) != 2 {
			continue
		}

		valuesList, found := parsedTags[pair[0]]
		if !found {
			parsedTags[pair[0]] = make([]string, 0)
		}
		valuesList = append(valuesList, pair[1])
		parsedTags[pair[0]] = valuesList
	}

	return parsedTags
}

func isValidRegion(region string) bool {
	_, found := VALID_REGIONS[region]
	return found
}

func parseFlags() options {
	viper.SetConfigName(".noclickops")
	// paths are searched in the order they've been added
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/noclickops/")
	viper.AddConfigPath("/etc/noclickops/")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error when attempting to find and read the config file %w", err)
	}

	var opts options

	flag.StringVar(&opts.stateFile, "statefile", "", "The statefile to parse")
	flag.StringVar(&opts.s3Bucket, "s3-bucket", "", "Download statefile(s) from this s3 bucket")
	flag.StringVar(&opts.s3BucketRegion, "s3-bucket-region", "", "The bucket's region")
	flag.StringVar(&opts.regions, "regions", "", "Comma-separated list of regions to check")
	flag.StringVar(&opts.ignoreTags, "ignore-tags", "", "Comma-separated list of 'key=value' pairs of tags to ignore (e.g. mytag=myvalue,another/tag=another-value)")
	flag.BoolVar(&opts.removeDownloadedStatefiles, "remove-downloaded-statefiles", false, "If specified, any downloaded statefiles will be deleted at the end")
	flag.BoolVar(&opts.forceDownload, "force-download", false, "If specified, it will download all the files from the bucket even they overwrite existing ones")
	flag.Parse()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	log.Fatal(viper.Get("s3-bucket"))
	err = opts.validate()
	if err != nil {
		log.Fatal(err)
	}

	return opts
}
