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
	EKS_node_group
	SSOAdmin_permission_set
	Identitystore_user
	Identitystore_group
	Instance
	Eip
	DB_instance
	RDS_cluster
	SNS_topic
	SNS_subscription
	S3_bucket
	CloudFront_distribution
	ELB_load_balancer
	ELBV2_load_balancer
	VPC
	Internet_gateway
	NAT_gateway
	Subnet
	VPC_endpoint
	Autoscaling_group
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
	SNS
	S3
	CloudFront
	ELB
	ELBV2
	ASG
	ResourceGroupsTaggingAPI
)

type Resource struct {
	Arn          string       `json:"arn"`
	TerraformID  string       `json:"terraform_id"`
	ResourceType ResourceType `json:"resource_type"`
	Region       string       `json:"region"`
}

type FilteredMeta struct {
	Found        int     `json:"found"`
	Managed      int     `json:"managed"`
	Unmanaged    int     `json:"unmanaged"`
	PctUnmanaged float32 `json:"pct_unmanaged"`
}

type FilteredResults struct {
	Resources []Resource   `json:"resources"`
	Meta      FilteredMeta `json:"meta"`
}

type ServiceMeta struct {
	Global      bool
	ServiceName string
	AccountId   string
}

func (m ServiceMeta) GetServiceName() string { return m.ServiceName }

type NoclickopsService interface {
	GetAllResources() []Resource
	GetServiceName() string
}
