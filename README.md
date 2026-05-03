# NoClickOps

## What it is

Noclickops intents to be a simple tool to see which and how many of your AWS resources are captured
in terraform.

## What it isn't

1. Complete

It only checks the resources listed below. It's not and it has no intention of becoming a tool that
will list every single possible resource. It's meant to be informational, to help you find the
biggest gaps in your terraform coverage.

2. Replacement for terraform's [list](https://developer.hashicorp.com/terraform/language/block/tfquery/list)/[import](https://developer.hashicorp.com/terraform/language/block/import) tools

Noclickops' intention is to help you decide which resources should take precedence over others, and then use terraform's tools to list and import resources to your code.

3. Replacement for tools like [terraformer](https://github.com/GoogleCloudPlatform/terraformer/) 

Same as above, the intention is to guide your decisions, it is not and will not become a tool to generate terraform code.


## How it works

It reads either a local statefile or the statefiles stored in an S3 bucket, extracts all the IDs and
then goes through each of the supported services, lists the resources, builds the corresponding
terraform IDs (i.e. what you would have used in a `terraform import` statement), checks if they
exist in your statefiles, and prints the missing ones in a json format for ease of querying.

## Supported services and resources

|    Service    	|                                          Resources                                           	|
|:-------------:	|:--------------------------------------------------------------------------------------------:	|
|    route53    	|                              route53_zone, route53_record                                    	|
|      iam      	|                           iam_user, iam_group, iam_policy                                    	|
|      ssm      	|                                      ssm_parameter                                           	|
|      ec2      	|   security_group, security_group_rule, instance, eip, vpc, internet_gateway, nat_gateway,    	|
|               	|                                   subnet, vpc_endpoint                                       	|
|      eks      	|                             eks_cluster, eks_node_group                                      	|
| identitystore 	|                        identitystore_group, identitystore_user                               	|
|    ssoadmin   	|                                 ssoadmin_permission_set                                      	|
|      rds      	|                                 db_instance, rds_cluster                                     	|
|      sns      	|                              sns_topic, sns_subscription                                     	|
|      s3       	|                                        s3_bucket                                             	|
|  cloudfront   	|                                  cloudfront_distribution                                     	|
|      elb      	|                                      elb_load_balancer                                       	|
|     elbv2     	|                                     elbv2_load_balancer                                      	|
| autoscaling   	|                                    autoscaling_group                                         	|
|    lambda     	|                                   lambda_function                                            	|

## How to use

### Use with a local statefile and a single region

```
noclickops --statefile ./example.tfstate --regions eu-west-1
```

### Use with an s3 bucket and multiple regions

```
noclickops --s3-bucket example-s3-statefile-bucket --s3-bucket-region eu-west-2 --regions eu-west-2,eu-west-1
```

### All flags

| Flag | Short | Description |
|---|---|---|
| `--statefile` | `-s` | Path to a local statefile to parse |
| `--s3-bucket` | `-b` | Download statefile(s) from this S3 bucket |
| `--s3-bucket-region` | `-k` | The region of the S3 bucket |
| `--regions` | `-r` | Comma-separated list of AWS regions to check |
| `--ignore-tags` | `-i` | Can be specified multiple times; each value is a `tagKey=value1,value2` pair - resources carrying any of these tags are excluded from results |
| `--delete-downloaded-state-files` | `-d` | Delete any statefiles downloaded from S3 when done |
| `--force-download` | `-f` | Re-download all files from S3 even if they already exist locally |

### Using a config file

Instead of passing flags every time, you can store your configuration in a `.noclickops.yml` file. Noclickops searches for this file in the following locations (in order):

1. Current directory (`./.noclickops.yml`)
2. `$HOME/.config/noclickops/.noclickops.yml`
3. `/etc/noclickops/.noclickops.yml`

Config file keys match the long flag names. The `regions` accepts a list of strings while
`ignore-tags` accepts a list where each entry is a map of a tag key to its allowed values:

```yaml
s3-bucket: my-terraform-states-bucket
s3-bucket-region: eu-west-2
regions:
  - eu-west-2
  - eu-west-1
  - eu-central-1

ignore-tags:
  - environment:
      - sandbox
      - staging
  - team:
      - platform
```

You can use the command line to override the configuration from the file.

## Ignoring resources by tag

Resources tagged with certain keys are automatically excluded from results without needing `--ignore-tags`:

- Tags whose key starts with `kubernetes.io/cluster/` (EKS-managed resources)
- Tags whose key starts with `aws:eks:` (EKS-managed resources)
- Tags whose key starts with `k8s.io/` (EKS-managed resources)
- Tags whose key starts with `aws:cloudformation:stack-name` (resources provisioned by cloudformation)
- Tags whose key starts with `noclickops/ignore`

To exclude additional resources, pass their tags with `--ignore-tags`:

```
# Ignore all resources tagged environment=sandbox
noclickops --statefile ./example.tfstate --regions eu-west-1 --ignore-tags environment=sandbox

# Multiple tag filters (any match causes the resource to be ignored)
noclickops --statefile ./example.tfstate --regions eu-west-1 --ignore-tags environment=sandbox,team=platform

# Multiple values for the same key
noclickops --statefile ./example.tfstate --regions eu-west-1 --ignore-tags environment=sandbox,environment=staging
```

## Output format

`noclickops` prints a JSON object with two top-level keys: `results` (per-service breakdown) and `summary` (account-wide totals).

```json
{
  "results": {
    "<service name>": {
      "resources": [
        {
          "arn": "<the arn of the resource>",
          "terraform_id": "<the id you would have used in `terraform import`, which varies per resource>",
          "resource_type": "<the resource name>",
          "region": "<global | specific region>"
        }
      ],
      "meta": {
        "found": "<total number of resources found in AWS>",
        "managed": "<number of resources found in statefiles>",
        "unmanaged": "<number of resources not found in statefiles>",
        "ignored": "<number of resources excluded due to tags>",
        "pct_unmanaged": "<percentage of unmanaged resources>"
      }
    }
  },
  "summary": {
    "found_in_aws": "<total resources found across all services>",
    "found_in_terraform": "<total managed resources>",
    "not_found_in_terraform": "<total unmanaged resources>",
    "ignored": "<total ignored resources>",
    "pct_unmanaged": "<overall unmanaged percentage>"
  }
}
```

Example:

```json
{
  "results": {
    "iam": {
      "resources": [
        {
          "arn": "arn:aws:iam::1234567890:user/admin",
          "terraform_id": "admin",
          "resource_type": "iam_user",
          "region": "global"
        },
        {
          "arn": "arn:aws:iam::1234567890:user/sakisv",
          "terraform_id": "sakisv",
          "resource_type": "iam_user",
          "region": "global"
        },
        {
          "arn": "arn:aws:iam::1234567890:group/admins",
          "terraform_id": "admins",
          "resource_type": "iam_group",
          "region": "global"
        }
      ],
      "meta": {
        "found": 10,
        "managed": 7,
        "unmanaged": 3,
        "ignored": 0,
        "pct_unmanaged": 30.0
      }
    },
    "ssm": {
      "resources": [
        {
          "arn": "arn:aws:ssm:eu-west-2:1234567890:parameter/an/example/ssm/parameter",
          "terraform_id": "/an/example/ssm/parameter",
          "resource_type": "ssm_parameter",
          "region": "eu-west-2"
        }
      ],
      "meta": {
        "found": 5,
        "managed": 4,
        "unmanaged": 1,
        "ignored": 0,
        "pct_unmanaged": 20.0
      }
    }
  },
  "summary": {
    "found_in_aws": 15,
    "found_in_terraform": 11,
    "not_found_in_terraform": 4,
    "ignored": 0,
    "pct_unmanaged": 26.67
  }
}
```

### Saving locally for quicker analysis

To avoid having to query AWS repeatedly, redirect the output to a file:

```
noclickops --s3-bucket example-s3-statefile-bucket --s3-bucket-region eu-west-2 --regions eu-west-2,eu-west-1 > unmanaged_resources.json
```


## Using `jq` to narrow down the list

You can pipe the output directly to `jq` or store it in a file first. The examples below use the file approach.

### Find which services have unmanaged resources

```
cat unmanaged_resources.json | jq '.results | keys'
```

### List unmanaged resources from a specific service only

```
cat unmanaged_resources.json | jq '.results.ssm.resources'
```

### List specific resource types

```
cat unmanaged_resources.json | jq '.results[] | .resources[] | select(.resource_type | contains("iam_policy"))'
```

or, if you know the service:

```
cat unmanaged_resources.json | jq '.results.iam.resources[] | select(.resource_type | contains("iam_policy"))'
```


### List all the different resource types

```
cat unmanaged_resources.json | jq '[.results[].resources[] | .resource_type] | unique'
```

Alternatively, if you also want counts:

```
cat unmanaged_resources.json | jq '.results[] | .resources[].resource_type' | sort | uniq -c
```

### Show the unmanaged percentage per service

```
cat unmanaged_resources.json | jq '.results | to_entries[] | {service: .key, pct_unmanaged: .value.meta.pct_unmanaged}'
```

### Show the overall summary

```
cat unmanaged_resources.json | jq '.summary'
```
