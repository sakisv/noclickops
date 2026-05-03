package main

import (
	"errors"
	"fmt"
	"log"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

type ConfigValues int

const (
	Statefile ConfigValues = iota
	S3Bucket
	S3BucketRegion
	DeleteDownloadedStateFiles
	ForceDownload
	Regions
	IgnoreTags
)

var configValues = map[ConfigValues]string{
	Statefile:                  "statefile",
	S3Bucket:                   "s3-bucket",
	S3BucketRegion:             "s3-bucket-region",
	DeleteDownloadedStateFiles: "delete-downloaded-state-files",
	ForceDownload:              "force-download",
	Regions:                    "regions",
	IgnoreTags:                 "ignore-tags",
}

type NoclickopsConfig struct {
	stateFile                  string
	deleteDownloadedStatefiles bool
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

	pflag.StringP(configValues[Statefile], "s", "", "The statefile to parse")
	pflag.StringP(configValues[S3Bucket], "b", "", "Download statefile(s) from this s3 bucket")
	pflag.StringP(configValues[S3BucketRegion], "k", "", "The bucket's region")
	pflag.StringP(configValues[Regions], "r", "", "Comma-separated list of regions to check")
	pflag.StringArrayP(configValues[IgnoreTags], "i", []string{}, "Can be used multiple times to provide list of 'tagKey=value1,value2' tags to ignore")
	pflag.BoolP(configValues[DeleteDownloadedStateFiles], "d", false, "If specified, any downloaded statefiles will be deleted at the end")
	pflag.BoolP(configValues[ForceDownload], "f", false, "If specified, it will download all the files from the bucket even they overwrite existing ones")
	pflag.Parse()

	// use this to pass control to viper
	// This means that any references to `statefile` will be resolved in the
	// right order (i.e. default, config, env, commandline)
	viper.BindPFlags(pflag.CommandLine)

	config := NewConfig(viper.GetViper())

	err = config.validate()
	if err != nil {
		log.Fatal(err)
	}

	return config
}
