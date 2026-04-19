package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/noclickops/common"
)

type ClientMeta struct {
	Region string
}

var SERVICES = map[common.AWSServiceName]common.ServiceMeta{
	common.IAM:        {Global: true, ServiceName: "iam"},
	common.Route53:    {Global: true, ServiceName: "route53"},
	common.CloudFront: {Global: true, ServiceName: "cloudfront"},

	common.EKS:           {Global: false, ServiceName: "eks"},
	common.IdentityStore: {Global: false, ServiceName: "identitystore"},
	common.SSM:           {Global: false, ServiceName: "ssm"},
	common.SSOAdmin:      {Global: false, ServiceName: "ssoadmin"},
	common.EC2:           {Global: false, ServiceName: "ec2"},
	common.RDS:           {Global: false, ServiceName: "rds"},
	common.SNS:           {Global: false, ServiceName: "sns"},
	common.S3:            {Global: false, ServiceName: "s3"},
	common.ELB:           {Global: false, ServiceName: "elb"},
	common.ELBV2:         {Global: false, ServiceName: "elbv2"},
	common.ASG:           {Global: false, ServiceName: "autoscaling"},

	// not included in the switch/case in `NewclickopsServiceFromConfigs`
	// because it doesn't follow the same invocation pattern.
	// It's only included here for convenience
	common.ResourceGroupsTaggingAPI: {Global: false, ServiceName: "resourcegroupstaggingapi"},
}

func NewNoclickopsServiceFromConfigs(service common.AWSServiceName, configs []aws.Config, accountId string) common.NoclickopsService {
	if len(configs) == 0 {
		panic("Cannot create clients without config")
	}

	meta, found := SERVICES[service]
	meta.AccountId = accountId
	if !found {
		panic("unknown service")
	}

	if meta.Global {
		configs = configs[:1]
	}

	switch service {
	case common.IAM:
		c := NewIAMServiceFromConfigs(configs, meta)
		return &c
	case common.Route53:
		c := NewRoute53ServiceFromConfigs(configs, meta)
		return &c
	case common.EKS:
		c := NewEKSServiceFromConfigs(configs, meta)
		return &c
	case common.SSM:
		c := NewSSMServiceFromConfigs(configs, meta)
		return &c
	case common.SSOAdmin:
		c := NewSSOAdminServiceFromConfigs(configs, meta)
		return &c
	case common.EC2:
		c := NewEC2ServiceFromConfigs(configs, meta)
		return &c
	case common.RDS:
		c := NewRDSServiceFromConfigs(configs, meta)
		return &c
	case common.S3:
		c := NewS3ServiceFromConfigs(configs, meta)
		return &c
	case common.SNS:
		c := NewSNSServiceFromConfigs(configs, meta)
		return &c
	case common.CloudFront:
		c := NewCloudFrontServiceFromConfigs(configs, meta)
		return &c
	case common.ELB:
		c := NewELBServiceFromConfigs(configs, meta)
		return &c
	case common.ELBV2:
		c := NewELBV2ServiceFromConfigs(configs, meta)
		return &c
	case common.ASG:
		c := NewAutoscalingServiceFromConfigs(configs, meta)
		return &c
	case common.IdentityStore:
		ssoMeta := SERVICES[common.SSOAdmin]
		ssoClient := NewSSOAdminServiceFromConfigs(configs[:1], ssoMeta)
		c := NewIdentityStoreServiceFromConfigs(configs, meta, &ssoClient)
		return &c
	}
	panic("unknown service")
}
