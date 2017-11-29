# Changelog

## 1.1.0
* Feature - Add support for Task Networking
* Feature - Add support for AWS Fargate
* Feature - Add support for viewing Cloudwatch logs of an ECS task
* Enhancement - Added Amazon Open Source Code of Conduct
* Bug - Fix credential resolution using aws-profile #369

## 1.0.0
* Feature - Add support for configuring multiple named ECS Profiles and Cluster Configurations #364
* Feature - Update the Order of Resolution for Credentials and Region #351
* Feature - Add support for Task Role Arn, Essential, and Network Mode with the —ecs-params flag #328
* Feature - Add support for running the ECS CLI on Windows #354
* Enhancement - Make keypair optional in `ecs-cli up` command #347
* Enhancement - Update CloudFormation template with all available instance types #340
* Bug - Change default cluster MinSize to 0 #336

## 0.6.6
* Feature - Add support for region ap-northeast-2

## 0.6.5
* Feature - Add support for cap_add and cap_drop
* Feature - Update AMI to amzn-ami-2017.03.g-amazon-ecs-optimized
* Enhancement - PS command displays private IPs when instance lacks a Public IP
* Bug - All commands now return an error exit code for error cases #306

## 0.6.4
* Feature - Update AMI to amzn-ami-2017.03.f-amazon-ecs-optimized

## 0.6.3
* Feature - Update AMI to amzn-ami-2017.03.e-amazon-ecs-optimized
* Feature - Support configurable timeout using new `--timeout` flag in `ecs-cli compose service` commands.
* Enhancement - Print service events when `ecs-cli compose service up` is run
* Feature - Support custom instance role by `--instance-role` flag in `ecs-cli up` command.


## 0.6.2
* Enhancement - Support region cn-north-1

## 0.6.1
* Enhancement - Support multiple compose files in compose commands
* Enhancement - Support `docker-compose.override.yml` with compose commands
* Bug - `--cluster` and `--region` flags can be specified both before and after compose and compose service subcommands

## 0.6.0
* Feature - Update ami to amzn-ami-2017.03.c-amazon-ecs-optimized
* Feature - Support cluster and region runtime flag for all ECS commands
* Feature - Support `--task-role-arn` in compose commands
* Feature - Support memory reservation in compose
* Feature - `ecs-cli up` without auto-assigned IP address
* Enhancement - Support Multiple Security Groups in the `ecs-cli up`
* Enhancement - Support `ecs-cli compose run` with multiple containers and run command overrides
* Enhancement - Support additional instance types p2, g2, and x1
* Bug - Avoid SIGSEGV error when ec2InstanceID does not exist #231
* Bug - Allow dashes “-“ in `ecs-cli push` #238
* Bug - Allow `ecs-cli compose up` to have project name longer than 36 characters #97

## 0.5.0
* Feature - Support ECR push, pull, and list images
* Feature - Support existing ELB/ALB in CreateService
* Feature - Update ami to amzn-ami-2016.09.g-amazon-ecs-optimized
* Enhancement - Added r4 instance types
* Bug - Add prompt to delete cluster [#186](https://github.com/aws/amazon-ecs-cli/pull/186)
* Bug - Creates new volume when there's no host path [#201](https://github.com/aws/amazon-ecs-cli/pull/201)
* Bug - `ecs-cli configure` truncates the file to avoid messing up the config file [#216](https://github.com/aws/amazon-ecs-cli/pull/216)

## 0.4.6
* Feature - Update ECS-optimized AMIs to latest 2016.09.d
* Bug - Support human readable strings for mem_limit
* Feature - Support for reading regions from aws profile
* Feature - Support for assume role from aws profile

## 0.4.5
* Feature - Update ECS-optimized AMIs to latest 2016.09.c
* Bug - When environment variable is not resolved, set it to empty string.
* Bug - `ecs-cli up` security group, vpc, subnets, azs validations
* Bug - Add `--force` flag to `ecs-cli up` to delete CloudFormation stack if it exists

## 0.4.4
* Feature - Update ECS-optimized AMIs to latest 2016.03.i.
* Bug - Add validation for cluster name in `ecs-cli up` command.

## 0.4.3
* Feature - Update ECS-optimized AMIs to latest 2016.03.h.
* Feature - Add support for different volumes_from format supported by Docker compose.

## 0.4.2
* Feature - Update ECS-optimized AMIs to latest 2016.03.f.
* Bug - Ensure least privilege for ~/.ecs/config file with permissions 0600.

## 0.4.1
* Feature - Update ECS-optimized AMIs to latest 2016.03.e.
* Bug - Fix `project-name` option for `ecs-cli compose` command to accept `-` in the name.

## 0.4.0
* Feature - Add support for `services` defined in the [Compose v2 file format](https://docs.docker.com/compose/compose-file/#/version-2).
* Feature - Add support for [variable substitution](https://docs.docker.com/compose/compose-file/#variable-substitution)
  in Compose files.
* Feature - Add support for [default environment file](https://docs.docker.com/compose/env-file/)
  `.env` placed in the folder `ecs-cli compose` command is executed from (current working directory).
* Bug - Fix several YAML parsing issues (with single quotes, JSON arrays, indentation issues)

## 0.3.1
* Feature - Update ECS-optimized AMIs to latest 2016.03.d.
* Bug - Fix issue to read credentials/role from EC2 instance metadata.

## 0.3.0
* Feature - Add support for compose option `env_file`.
* Feature - Add support for session environment variables for compose option
  `env_file` and `environment`.
* Feature - Add support for deployment parameters to compose service commands.
  Users can supply --deployment-max-percent and --deployment-min-healthy-percent to
  `ecs-cli compose service create/up/scale` commands
* Feature - Add support for configurable prefixes for resources created by the cli.
  Users can now call `ecs-cli configure` to configure
 * prefix used for the Cloudformation stack in `ecs-cli up` command,
 * compose project name prefix used for task definition and started by field
  in `ecs-cli compose` commands,
 * compose service name prefix used by `ecs-cli compose service` command
* Feature - Update ECS-optimized AMIs to latest 2016.03.a.
* Enhancement - Add License file to the ecs-cli executable. Users can view the License
  for the ECS CLI and its dependencies by calling `ecs-cli license`
* Enhancement - Update go-ini/ini to v1.11.0 and aws/aws-sdk-go to v1.1.14

## 0.2.1
* Feature - Update ECS-optimized AMIs to latest 2015.09.f

## 0.2.0
* Feature - Add support for new docker options in compose yaml file.
* Feature - Add new options to ecs-cli up (--image-id, --debug or --verbose).
* Feature - Add support for m4, d2, g2 instance types.
* Feature - Add new regions eu-central-1 and ap-southeast-1.
* Feature - Update ECS-optimized AMIs to latest 2015.09.e
  (with Amazon ECR support).
* Enhancement - Better error messaging for ecs-cli up and
  RegisterTaskDefinition API.
* Bug - Include region, account in key for local Task Definition cache.
* Bug - Change ordering of AWS Credential resolution for the ecs-cli.
* Bug - Minor bug fixes to CFN template (remove additional parameter from
  autoscaling creation, add internet gateway attachment dependency to public
  route)
