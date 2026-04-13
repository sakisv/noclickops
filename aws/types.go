package aws

import (
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/noclickops/common"
)

type ClientMeta struct {
	Region string
}

var SERVICES = map[common.AWSServiceName]common.ServiceMeta{
	common.IAM:     {Global: true, ServiceName: "iam"},
	common.Route53: {Global: true, ServiceName: "route53"},

	common.EKS:            {Global: false, ServiceName: "eks"},
	common.IdentityStore:  {Global: false, ServiceName: "identitystore"},
	common.SSM:            {Global: false, ServiceName: "ssm"},
	common.SSOAdmin:       {Global: false, ServiceName: "ssoadmin"},
	common.SecurityGroups: {Global: false, ServiceName: "securitygroups"},
}

func NewNoclickopsServiceFromConfigs(service common.AWSServiceName, configs []aws.Config) common.NoclickopsService {
	if len(configs) == 0 {
		panic("Cannot create clients without config")
	}

	meta := SERVICES[service]
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
	case common.SecurityGroups:
		c := NewEC2ServiceFromConfigs(configs, meta)
		return &c
	case common.IdentityStore:
		ssoMeta := SERVICES[common.SSOAdmin]
		ssoClient := NewSSOAdminServiceFromConfigs(configs[:1], ssoMeta)
		c := NewIdentityStoreServiceFromConfigs(configs, meta, &ssoClient)
		return &c
	}
	panic("unknown service")
}
