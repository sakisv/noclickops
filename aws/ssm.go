package aws

import (
	"context"
	"fmt"
	"log"
	"strconv"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/ssm"
)

func GetAllParametersNames(cfg aws.Config) []string {
	ssm_client := ssm.NewFromConfig(cfg)

	res, err := ssm_client.GetParametersByPath(context.TODO(), &ssm.GetParametersByPathInput{
		Path:       aws.String("/"),
		MaxResults: aws.Int32(10),
		Recursive:  aws.Bool(true),
	})

	if err != nil {
		log.Fatal(err)
	}
	var ids []string
	for _, el := range res.Parameters {
		ids = append(ids, *el.Name)
	}
	fmt.Println("Found " + strconv.Itoa(len(ids)) + " parameters")
	return ids
}
