# Changelog

## 1.12.0
* Feature - Add support for IPC and PID Docker flags #669

## 1.11.1
* Bug - Revert IPC/PID flags due to bad default behavior

## 1.11.0
* Feature - Add support ECS Secrets for #664
* Feature - Add support for IPC and PID flags #665
* Feature - Add support for mandatory variables in docker-compose #651
* Enhancement - Add support for FIPs endpoint when using ECR #666

## 1.10.0
* Feature - Add `registry-creds` command as part of Private Registry Authentication workflow #652 #601
* Feature - Use Amazon Linux 2 ECS Optimized AMI #647
* Bug - Catch errors from missing user data files #646
* Bug - Fix '--create-log-groups' to leave region for Service Discovery resources unchanged #644

## 1.9.0
* Feature - Add support for Service Discovery (#485)
* Feature - Add support for EC2 Spot Instances in ECS Clusters (#396)
* Feature - Add support for custom user data (#16)
* Bug - Fix error using env vars with nil value (#620)
* Enhancement - Improve `logs` command behavior and error handling (#612)
* Enhancement - Add support for GO 1.11 (#632)
* Enhancement - Add support for new EC2 Instance types (#630, #618)

## 1.8.0
* Feature - Add support for volumes with docker volume configuration in ECS Params #587
* Feature - Add support for task placement constraints and strategies in ECS Params (#515, #212)
* Feature - Add `--force-update` on `compose up` to force re-creation of tasks
* Feature - Add support for specifying Private Registry Authentication credentials in ECS Params #573

## 1.7.0
* Feature - Add support for container health check (#472)
* Feature - Add support for devices (#508)
* Bug - Fix error in ps command (#522)
* Bug - Fix error using ENV variables with docker compose v3 (#537)
* Bug - Fix memory validation in containers (#546)
* Bug - Fix log message for container resource overrides
* Bug - Add missing cn-northwest-1 region in Cloudformation template (#552)
* Enhancement - Add waiter for service creation (#79)

## 1.6.0
* Feature - Add support for docker Compose file version 3 (#218)
* Feature - Add support for environmental variables in ecs-params.yml (#530)
* Feature - Add support for named volumes (#481)
* Bug - Fix support for slashes in image names (#361)
* Bug - Fix stack timeout message for CFN stack deletion
* Bug - Fix exit code to be 1 for all CLI usage errors (#490)
* Enhancement - Add Pull Request template (#492)

## 1.5.0
* Feature - Add support for tmpfs
* Feature - Add support for shm_size
* Feature - Add Amazon ECS PGP Public Key and instructions on verifying signatures
* Feature - Retrieve ECS AMI ID from SSM on cluster creation

## 1.4.2
* Feature - Update AMI to amzn-ami-2017.09.k-amazon-ecs-optimized

## 1.4.1
* Bug - Ensure tests pass on go 1.10
* Enhancement - Support longer resource IDs in Cloudformation template

## 1.4.0
* Feature - Update AMI to amzn-ami-2017.09.j-amazon-ecs-optimized
* Feature - Add force-deployment flag to compose service (#144)
* Feature - Support aws_session token in ECS Profiles (#415)
* Feature - Add support for us-gov-west-1
* Bug - Fix YAML parse warnings on networks field (#237)
* Enhancement - Add issue template

## 1.3.0
* Feature - Update AMI to amzn-ami-2017.09.g-amazon-ecs-optimized
* Feature - Add health-check-grace-period flag for compose service up
* Feature - Add empty flag for cluster up

## 1.2.2
* Feature - Update AMI to amzn-ami-2017.09.f-amazon-ecs-optimized

## 1.2.1
* Feature - Update AMI to amzn-ami-2017.09.e-amazon-ecs-optimized

## 1.2.0
* Feature - Added `--create-log-groups` flag to create the CloudWatch log groups specified in your compose file. #389
* Feature - Add support for region ap-south-1, sa-east-1, and eu-west-3
* Enhancement - Update CloudFormation template with all available instance types #379
* Enhancement - Make `ecs-cli scale` compatible with CloudFormation Templates created by the ECS Console #390
* Bug - Fixed `ecs-cli up` with EC2 Launch Type and a custom instance role #394
* Bug - Make `ecs-cli scale` compatible with CloudFormation templates created by older version of the ECS CLI #330

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
