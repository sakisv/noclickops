# NoClickOps

A tool to see which and how much of your AWS infrastructure is captured in terraform.

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

## How to use

### Use with a local statefile and a single region

```
noclickops ./example.tfstate --regions eu-west-1
```

### Use with an s3 bucket and multiple regions

```
noclickops -s3-bucket example-s3-statefile-bucket -s3-bucket-region eu-west-2 --regions eu-west-2,eu-west-1
```

## Output management

### Format

`noclickops` prints the missing resources in the following json format:

```json
{
  "<service name>": {
    "resources": [
      {
        "terraform_id": "<the id you would have used in `terraform import` which varies per resource>",
        "resource_type": "<the resource name>",
        "region": "<global | specific region>"
      }
    ],
    "meta": {
      "found": "<total number of resources found in AWS>",
      "managed": "<number of resources found in statefiles>",
      "unmanaged": "<number of resources not found in statefiles>",
      "pct_unmanaged": "<percentage of unmanaged resources>"
    }
  }
}
```

Example:

```json
{
  "iam": {
    "resources": [
      {
        "terraform_id": "admin",
        "resource_type": "iam_user",
        "region": "global"
      },
      {
        "terraform_id": "sakisv",
        "resource_type": "iam_user",
        "region": "global"
      },
      {
        "terraform_id": "admins",
        "resource_type": "iam_group",
        "region": "global"
      }
    ],
    "meta": {
      "found": 10,
      "managed": 7,
      "unmanaged": 3,
      "pct_unmanaged": 30.0
    }
  },
  "ssm": {
    "resources": [
      {
        "terraform_id": "/an/example/ssm/parameter",
        "resource_type": "ssm_parameter",
        "region": "eu-west-2"
      }
    ],
    "meta": {
      "found": 5,
      "managed": 4,
      "unmanaged": 1,
      "pct_unmanaged": 20.0
    }
  }
}
```

### Saving locally for quicker analysis

To avoid having to query AWS if you just want to do a quick check of the findings, you can redirect the output to a file:

```
noclickops -s3-bucket example-s3-statefile-bucket -s3-bucket-region eu-west-2 --regions eu-west-2,eu-west-1 > unmanaged_resources.json
```


## Using `jq` to narrow down the list

You can use `jq` to slice the results in more manageable chunks either by piping the output directly to `jq` or by storing it in a file and then `cat`ing it.

The examples below use the second method, but they'd work on the first as well

### Find which services have unmanaged resources

```
cat unmanaged_resources.json | jq 'keys'
```

### List unmanaged resources from a specific service only

```
cat unmanaged_resources.json | jq '.ssm.resources'
```

### List specific resource types

```
cat unmanaged_resources.json | jq '.[] | .resources[] | select(.resource_type | contains("iam_policy"))'
```

or, if you know the service

```
cat unmanaged_resources.json | jq '.iam.resources[] | select(.resource_type | contains("iam_policy"))'
```


### List all the different resource types

```
cat unmanaged_resources.json | jq '[.[].resources[] | .resource_type] | unique'
```

Alternatively, if you also want counts:

```
cat unmanaged_resources.json | jq '.[] | .resources[].resource_type' | sort | uniq -c
```

### Show the unmanaged percentage per service

```
cat unmanaged_resources.json | jq 'to_entries[] | {service: .key, pct_unmanaged: .value.meta.pct_unmanaged}'
```
