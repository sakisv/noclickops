package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/iam"
	"github.com/aws/aws-sdk-go-v2/service/route53"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
	claws "github.com/noclickops/aws"
	"github.com/noclickops/common"
)

const STATEFILES_DIR = ".cache/noclickops/statefiles/"

func getManagedIDs(statefile_paths []string) map[string]struct{} {
	reg := regexp.MustCompile(`\"id\": \".*\",?`)

	var managed_ids = make(map[string]struct{})
	for _, path := range statefile_paths {
		contents_b, err := os.ReadFile(path)
		if err != nil {
			log.Fatal(err)
		}
		contents := string(contents_b[:])
		finds := reg.FindAllString(contents, -1)
		fmt.Println()

		for _, el := range finds {
			res := strings.Split(el, "\": ")
			if len(res) != 2 {
				continue
			}
			managed_id, _ := strings.CutSuffix(res[1], ",")
			managed_id = strings.ReplaceAll(managed_id, "\"", "")
			_, ok := managed_ids[managed_id]
			if !ok {
				managed_ids[managed_id] = struct{}{}
			}
		}
	}
	return managed_ids
}

func filter(managedIds map[string]struct{}, foundRecords map[string][]common.Resource) map[string][]common.Resource {
	unmanagedResourceIds := make(map[string][]common.Resource)
	for key, value := range foundRecords {
		if len(value) == 0 {
			continue
		}
		for _, el := range value {
			_, found := managedIds[el.TerraformID]
			if found {
				//println("[DEBUG] Found " + el)
			} else {
				unmanagedResourceIds[key] = append(unmanagedResourceIds[key], el)
				//println("[DEBUG] Not found " + el)
			}
		}
	}
	return unmanagedResourceIds
}

func getFullPathToHomeTarget(to string) string {
	homedir, _ := os.UserHomeDir()
	return path.Join(homedir, to)
}

func createRemoteStatefilesDir() string {
	target_dir := getFullPathToHomeTarget(STATEFILES_DIR)
	err := os.MkdirAll(target_dir, 0755)
	if err != nil {
		log.Fatal(err)
	}
	return target_dir
}

func delete_statefiles_dir() error {
	target_dir := getFullPathToHomeTarget(STATEFILES_DIR)
	//println("[DEBUG] Deleting " + target_dir)
	return os.RemoveAll(target_dir)
}

func download_statefiles_from_s3(bucket string, cfg aws.Config) []string {
	client := s3.NewFromConfig(cfg)
	var files []string
	res, err := client.ListObjectsV2(context.TODO(), &s3.ListObjectsV2Input{
		Bucket: aws.String(bucket),
	})
	if err != nil {
		log.Fatal(err)
	}
	statefiles_dir := createRemoteStatefilesDir()

	for _, object := range res.Contents {
		filename := strings.ReplaceAll(*object.Key, "/", "~~")
		full_path := path.Join(statefiles_dir, filename)
		//println("[DEBUG] Downloading " + *object.Key + " into " + full_path)

		getObjectResp, err := client.GetObject(context.TODO(), &s3.GetObjectInput{
			Bucket: aws.String(bucket),
			Key:    object.Key,
		})
		if err != nil {
			log.Fatal(err)
		}
		defer getObjectResp.Body.Close()

		file, err := os.Create(full_path)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		body, err := io.ReadAll(getObjectResp.Body)

		_, err = file.Write(body)
		if err != nil {
			log.Fatal(err)
		}

		files = append(files, full_path)
	}

	return files
}

func main() {
	var stateFile string
	var region string
	var s3_bucket string
	flag.StringVar(&stateFile, "statefile", "", "The statefile to parse")
	flag.StringVar(&s3_bucket, "s3_bucket", "", "Download statefile(s) from this s3 bucket")
	flag.StringVar(&region, "region", "eu-west-1", "The AWS region to target")
	flag.Parse()

	if stateFile == "" && s3_bucket == "" {
		fmt.Println("At least one of s3_bucket or statefile must be provided")
		fmt.Println("Use -h / --help")
		return
	}

	cfg, err := config.LoadDefaultConfig(context.TODO())
	if err != nil {
		log.Fatal(err)
	}

	downloaded_files := download_statefiles_from_s3(s3_bucket, cfg)
	if stateFile != "" {
		downloaded_files = append(downloaded_files, stateFile)
	}
	managedIDs := getManagedIDs(downloaded_files)
	defer delete_statefiles_dir()

	foundRecords := make(map[string][]common.Resource)
	foundRecords["policies"] = claws.GetAllPoliciesArns(iam.NewFromConfig(cfg))
	foundRecords["ssm_params"] = claws.GetAllParametersNames(ssm.NewFromConfig(cfg))
	foundRecords["route53_records"] = claws.GetAllRoute53RecordIds(route53.NewFromConfig(cfg))

	unmanagedResourceIds := filter(managedIDs, foundRecords)
	json, err := json.Marshal(unmanagedResourceIds)
	fmt.Println(string(json))

}
