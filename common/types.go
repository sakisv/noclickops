package common

type ResourceType int

const (
	Route53_zone ResourceType = iota
	Route53_record
	IAM_policy
	SSM_parameter
)

type Resource struct {
	TerraformID  string
	ResourceType ResourceType
}
