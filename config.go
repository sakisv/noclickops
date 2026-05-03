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
	ignoreTags                 []string
	regionsList                []string
	ignoreTagsMap              map[string][]string
}

func NewConfig(v *viper.Viper) NoclickopsConfig {
	var config = NoclickopsConfig{}
	config.stateFile = v.GetString(configValues[Statefile])
	config.s3Bucket = v.GetString(configValues[S3Bucket])
	config.s3BucketRegion = v.GetString(configValues[S3BucketRegion])
	config.deleteDownloadedStatefiles = v.GetBool(configValues[DeleteDownloadedStateFiles])
	config.forceDownload = v.GetBool(configValues[ForceDownload])
	config.regionsList = v.GetStringSlice(configValues[Regions])

	switch t := v.Get(configValues[Regions]).(type) {
	// if received from flag, it will be a string
	case string:
		config.regions = t
		config.regionsList = strings.Split(config.regions, ",")
	// if read from file, it will be []interface{}
	case []any:
		for _, region := range t {
			s, ok := region.(string)
			if !ok {
				fmt.Printf("could not convert '%v' to string for %v", region, configValues[Regions])
			}
			config.regionsList = append(config.regionsList, s)
		}
		config.regions = strings.Join(config.regionsList, ",")
	default:
		fmt.Printf("unexpected '%v' type %T", configValues[Regions], t)
	}

	switch t := v.Get(configValues[IgnoreTags]).(type) {
	// if received from flag, it will be a []string
	case []string:
		config.ignoreTags = t
		config.ignoreTagsMap = parseTags(t)
	// if read from file, it will be []interface{}, so we use Unmarshalkey to
	// cast into the structure we need
	case []any:
		v.UnmarshalKey(configValues[IgnoreTags], &config.ignoreTagsMap)
		for k, v := range config.ignoreTagsMap {
			tagString := fmt.Sprintf("%v=%v", k, strings.Join(v, ","))
			config.ignoreTags = append(config.ignoreTags, tagString)
		}
	default:
		fmt.Printf("unexpected '%v' type %T", configValues[IgnoreTags], t)
	}

	return config
}

func (config *NoclickopsConfig) validate() error {
	var errs []error

	if config.stateFile == "" && config.s3Bucket == "" {
		errs = append(errs, fmt.Errorf("At least one of 's3-bucket' or 'statefile' must be provided"))
	}

	if config.s3Bucket != "" {
		if config.s3BucketRegion == "" {
			errs = append(errs, fmt.Errorf("s3-bucket-region must be provided if s3-bucket is defined"))
		}
		if !isValidRegion(config.s3BucketRegion) {
			errs = append(errs, fmt.Errorf("'%v' is not a valid region", config.s3BucketRegion))
		}
	}

	if config.forceDownload && config.s3Bucket == "" {
		errs = append(errs, fmt.Errorf("'force-download' must be used alongside an 's3-bucket'"))
	}

	for region := range strings.SplitSeq(config.regions, ",") {
		r := strings.ToLower(strings.TrimSpace(region))
		if !isValidRegion(r) {
			errs = append(errs, fmt.Errorf("'%v' is not a valid region", r))
			continue
		}

		config.regionsList = append(config.regionsList, r)
	}

	config.ignoreTagsMap = parseTags(config.ignoreTags)

	if len(errs) > 0 {
		errs = append(errs, fmt.Errorf("Use -h / --help"))
		return errors.Join(errs...)
	}
	return nil
}

func parseTags(tags []string) map[string][]string {
	parsedTags := make(map[string][]string)

	for _, tag := range tags {
		parts := strings.SplitN(tag, "=", 2)
		if len(parts) != 2 {
			continue
		}
		parsedTags[parts[0]] = strings.Split(parts[1], ",")
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
