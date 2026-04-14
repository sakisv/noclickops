package common

import "strings"

type ResourceType int

//go:generate stringer -type=ResourceType
const (
	Route53_zone ResourceType = iota
	Route53_record
	IAM_policy
	IAM_user
	IAM_group
	SSM_parameter
	Security_group
	Security_group_rule
	EKS_cluster
	SSOAdmin_permission_set
	Identitystore_user
	Identitystore_group
	Instance
	Eip
	DB_instance
	RDS_cluster
)

func (r ResourceType) MarshalJSON() ([]byte, error) {
	return []byte(`"` + strings.ToLower(r.String()) + `"`), nil
}

type AWSServiceName int

const (
	Route53 AWSServiceName = iota
	IAM
	EKS
	SSM
	EC2
	RDS
	SSOAdmin
	IdentityStore
)

type Resource struct {
	TerraformID  string       `json:"terraform_id"`
	ResourceType ResourceType `json:"resource_type"`
	Region       string       `json:"region"`
}

type ServiceMeta struct {
	Global      bool
	ServiceName string
}

func (m ServiceMeta) GetServiceName() string { return m.ServiceName }

type NoclickopsService interface {
	GetAllResources() []Resource
	GetServiceName() string
}
