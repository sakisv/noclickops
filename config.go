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

type NoclickopsConfig struct {
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

func (config *NoclickopsConfig) validate() error {
	var errs []string

	if config.stateFile == "" && config.s3Bucket == "" {
		errs = append(errs, "At least one of 's3-bucket' or 'statefile' must be provided")
	}

	if config.s3Bucket != "" {
		if config.s3BucketRegion == "" {
			errs = append(errs, "s3-bucket-region must be provided if s3-bucket is defined")
		}
		if !isValidRegion(config.s3BucketRegion) {
			errs = append(errs, fmt.Sprintf("'%v' is not a valid region", config.s3BucketRegion))
		}
	}

	if config.forceDownload && config.s3Bucket == "" {
		errs = append(errs, "'force-download' must be used alongside an 's3-bucket'")
	}

	for region := range strings.SplitSeq(config.regions, ",") {
		r := strings.ToLower(strings.TrimSpace(region))
		if !isValidRegion(r) {
			errs = append(errs, fmt.Sprintf("'%v' is not a valid region", r))
			continue
		}

		config.regionsList = append(config.regionsList, r)
	}

	config.ignoreTagsMap = parseTags(config.ignoreTags)

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

func loadConfig() NoclickopsConfig {
	viper.SetConfigName(".noclickops")
	// paths are searched in the order they've been added
	viper.AddConfigPath(".")
	viper.AddConfigPath("$HOME/.config/noclickops/")
	viper.AddConfigPath("/etc/noclickops/")

	err := viper.ReadInConfig()
	if err != nil {
		log.Fatal("Error when attempting to find and read the config file %w", err)
	}

	var config NoclickopsConfig

	flag.StringVar(&config.stateFile, "statefile", "", "The statefile to parse")
	flag.StringVar(&config.s3Bucket, "s3-bucket", "", "Download statefile(s) from this s3 bucket")
	flag.StringVar(&config.s3BucketRegion, "s3-bucket-region", "", "The bucket's region")
	flag.StringVar(&config.regions, "regions", "", "Comma-separated list of regions to check")
	flag.StringVar(&config.ignoreTags, "ignore-tags", "", "Comma-separated list of 'key=value' pairs of tags to ignore (e.g. mytag=myvalue,another/tag=another-value)")
	flag.BoolVar(&config.removeDownloadedStatefiles, "remove-downloaded-statefiles", false, "If specified, any downloaded statefiles will be deleted at the end")
	flag.BoolVar(&config.forceDownload, "force-download", false, "If specified, it will download all the files from the bucket even they overwrite existing ones")
	flag.Parse()

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	log.Fatal(viper.Get("s3-bucket"))
	err = config.validate()
	if err != nil {
		log.Fatal(err)
	}

	return config
}
