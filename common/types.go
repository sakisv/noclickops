package common

type ResourceType int

const (
	Route53_zone ResourceType = iota
	Route53_record
	IAM_policy
	IAM_user
	IAM_group
	SSM_parameter
	EC2_securitygroup
	EC2_securitygrouprule
	EKS_cluster
	SSOAdmin_identitystoreinstance
	SSOAdmin_permissionset
	SSOAdmin_accountassignments
	IdentityStore_user
	IdentityStore_group
)

type AWSService int

const (
	Route53 AWSService = iota
	IAM
	EKS
	SSM
	SecurityGroups
	SSOAdmin
	IdentityStore
)

type Resource struct {
	TerraformID  string       `json:"terraform_id"`
	ResourceType ResourceType `json:"resource_type"`
}

type ServiceMeta struct {
	Global      bool
	ServiceName string
}

func (m ServiceMeta) GetServiceName() string { return m.ServiceName }

type ResourceFetcher interface {
	GetAllResources() []Resource
	GetServiceName() string
}
