# NoClickOps

A tool to see which and how much of your AWS infrastructure is captured in terraform.

## How it works

It reads either a local statefile or the statefiles stored in an S3 bucket, extracts all the IDs and
then goes through each of the supported services, lists the resources, builds the corresponding
terraform IDs (i.e. what you would have used in a `terraform import` statement), checks if they
exist in your statefiles, and prints the missing ones in a json format for ease of querying.

## Supported services and resources

|    Service    	|                Resources                	|
|:-------------:	|:---------------------------------------:	|
|    route53    	|       route53_zone, route53_record      	|
|      iam      	|     iam_user, iam_group, iam_policy     	|
|      ssm      	|              ssm_pararmeter             	|
|      ec2      	|   security_group, security_group_rule   	|
|      eks      	|               eks_cluster               	|
| identitystore 	| identitystore_group, identitystore_user 	|
|    ssoadmin   	|            ssoadmin_instances           	|

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
  "<service name>": [
    {
      "terraform_id": "<the id you would have used in `terraform import` which varies per resource>"
      "resource_type": "<the resource name>"
      "region": "<global | specific region>"
    }
  ]
}
```

Example:

```json
{
  "iam": [
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
  "ssm": [
    {
      "terraform_id": "/an/example/ssm/parameter",
      "resource_type": "ssm_parameter",
      "region": "eu-west-2"
    }
  ]
}
```

### Saving locally for quicker analysis

To avoid having to query AWS if you just want to do a quick check of the findings, you can redirect the output to a file:

```
noclickops -s3-bucket example-s3-statefile-bucket -s3-bucket-region eu-west-2 --regions eu-west-2,eu-west-1 > unmanaged_resources.json
```


## Using `jq` to narrow down the list

You can use `jq` to slice the results in more manageable chunks either by piping the output directly to `jq` or by storing it in a file and then `cat`ing .

The examples below use the second method, but they'd work on the first as well

### Find which services have unamanged resources

```
cat unmanaged_resources.json | jq 'keys'
```

### List unmanaged resources from a specific service only

```
cat unmanaged_resources.json | jq '.ssm'
```

### List specific resource types

```
cat unmanaged_resources.json | jq '.[] | .[] | select(.resource_type | contains("iam_policy"))'
```

or, if you know the service

```
cat unmanaged_resources.json | jq '.iam | .[] | select(.resource_type | contains("iam_policy"))'
```


### List all the different resource types

```
cat unmanaged_resources.json | jq '[.[].[] | .resource_type] | unique'
```

Alternatively, if you also want counts:

```
cat unmanaged_resources.json | jq '.[] | .[].resource_type' | sort | uniq -c
```
