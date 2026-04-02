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
	SSOAdmin_identitystoreinstance
	SSOAdmin_permissionset
	SSOAdmin_accountassignments
	IdentityStore_user
	IdentityStore_group
)

type Resource struct {
	TerraformID  string
	ResourceType ResourceType
}
