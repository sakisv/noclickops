package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"path"
	"regexp"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

const STATEFILES_DIR = ".cache/noclickops/statefiles/"

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
	return os.RemoveAll(target_dir)
}

func download_statefiles_from_s3(bucket string, forceDownload bool, cfg aws.Config) []string {
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
		if strings.HasSuffix(*object.Key, "/") {
			continue
		}
		filename := strings.ReplaceAll(*object.Key, "/", ".")
		full_path := path.Join(statefiles_dir, filename)

		_, err := os.Stat(full_path)
		fileExists := err == nil
		if fileExists && !forceDownload {
			files = append(files, full_path)
			continue
		}

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
		if err != nil {
			log.Fatal(err)
		}

		_, err = file.Write(body)
		if err != nil {
			log.Fatal(err)
		}

		files = append(files, full_path)
	}

	return files
}

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
