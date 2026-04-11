package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/noclickops/common"
)

var SERVICES = map[common.AWSService]common.ServiceMeta{
	common.IAM:     {Global: true, ServiceName: "iam"},
	common.Route53: {Global: true, ServiceName: "route53"},

	common.EKS:            {Global: false, ServiceName: "eks"},
	common.IdentityStore:  {Global: false, ServiceName: "identitystore"},
	common.SSM:            {Global: false, ServiceName: "ssm"},
	common.SSOAdmin:       {Global: false, ServiceName: "ssoadmin"},
	common.SecurityGroups: {Global: false, ServiceName: "securitygroups"},
}

func NewClientFromConfigs(service common.AWSService, configs []aws.Config) common.ResourceFetcher {
	if len(configs) == 0 {
		panic("Cannot create clients without config")
	}

	meta := SERVICES[service]
	if meta.Global {
		configs = configs[:1]
	}

	switch service {
	case common.IAM:
		c := NewIAMClientFromConfigs(configs, meta)
		return &c
	case common.Route53:
		c := NewRoute53ClientFromConfigs(configs, meta)
		return &c
	case common.EKS:
		c := NewEKSClientFromConfigs(configs, meta)
		return &c
	case common.SSM:
		c := NewSSMClientFromConfigs(configs, meta)
		return &c
	case common.SSOAdmin:
		c := NewSSOAdminClientFromConfigs(configs, meta)
		return &c
	case common.SecurityGroups:
		c := NewEC2ClientFromConfigs(configs, meta)
		return &c
	case common.IdentityStore:
		ssoMeta := SERVICES[common.SSOAdmin]
		ssoClient := NewSSOAdminClientFromConfigs(configs[:1], ssoMeta)
		c := NewIdentityStoreClientFromConfigs(configs, meta, &ssoClient)
		return &c
	}
	panic("unknown service")
}
