# Amazon ECS CLI

The Amazon ECS Command Line Interface (CLI) is a command line interface for Amazon Elastic Container
Service (Amazon ECS) that provides high-level commands to simplify creating, updating, and
monitoring clusters and tasks from a local development environment. The Amazon ECS CLI supports
[Docker Compose](https://docs.docker.com/compose/), a popular open-source tool for defining and
running multi-container applications. Use the CLI as part of your everyday development and testing
cycle as an alternative to the AWS Management Console.

For more information about Amazon ECS, see the [Amazon ECS Developer
Guide](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/Welcome.html).  For information
about installing and using the Amazon ECS CLI, see the [ECS Command Line
Interface](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI.html).

The AWS Command Line Interface (AWS CLI) is a unified client for AWS services that provides commands
for all public API operations. These commands are lower level than those provided by the Amazon ECS
CLI. For more information about supported services and to download the AWS CLI, see the [AWS Command
Line Interface](http://aws.amazon.com/cli/) product detail page.

- [Installing](#Installing)
	- [Latest version](#latest-version)
        - [Download Links for within China](#download-links-for-within-china)
	- [Download specific version](#)
- [Configuring the CLI](#configuring-the-cli)
- [Using the CLI](#using-the-cli)
	- [Creating an ECS Cluster](#creating-an-ecs-cluster)
	- [Starting/Running Tasks](#startingrunning-tasks)
	- [Creating a Service](#creating-a-service)
	- [Using ECS parameters](#using-ecs-parameters)
- [Amazon ECS CLI Commands](#amazon-ecs-cli-commands)
- [Contributing to the CLI](#contributing-to-the-cli)
- [License](#license)

## Installing

Download the binary archive for your platform, decompress the archive, and
install the binary on your `$PATH`. You can use the provided `md5` hash to
verify the integrity of your download.

### Latest version
* Linux:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-latest](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-latest)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-latest.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-latest.md5)
* Macintosh:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-latest](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-latest)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-latest.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-latest.md5)
* Windows:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-latest.exe](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-latest.exe)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-latest.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-latest.md5)

#### Download Links for within China

As of v0.6.2 the ECS CLI supports the cn-north-1 region in China. The following links are the exact same binaries, but they are localized within China to provide a faster download experience.

* Linux:
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-linux-amd64-latest](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-linux-amd64-latest)
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-linux-amd64-latest.md5](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-linux-amd64-latest.md5)
* Macintosh:
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-darwin-amd64-latest](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-darwin-amd64-latest)
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-darwin-amd64-latest.md5](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-darwin-amd64-latest.md5)
* Windows:
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-windows-amd64-latest.exe](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-windows-amd64-latest.exe)
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-windows-amd64-latest.md5](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-windows-amd64-latest.md5)

### Download specific version
Using the URLs above, replace `latest` with the desired tag, for example `v1.0.0`. After downloading, remember to rename the binary file to `ecs-cli`. Note that Windows is only supported starting with version `v1.0.0`.

* Linux:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v1.0.0](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v1.0.0)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v1.0.0.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v1.0.0.md5)
* Macintosh:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v1.0.0](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v1.0.0)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v1.0.0.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v1.0.0.md5)
* Windows:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-v1.0.0.exe](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-v1.0.0.exe)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-v1.0.0.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-v1.0.0.md5)

## Configuring the CLI

The Amazon ECS CLI requires some basic configuration information before you can use it, such as your
AWS credentials, the AWS region in which to create your cluster, and the name of the Amazon ECS
cluster to use. Configuration information is stored in the `~/.ecs` directory on macOS and Linux
systems and in `C:\Users\<username>\AppData\local\ecs` on Windows systems.

### ECS Profiles

The Amazon ECS CLI supports configuring multiple sets of AWS credentials as named profiles using the
`ecs-cli configure profile command`. These profiles can then be referenced when you run Amazon ECS
CLI commands using the `--ecs-profile` flag; if a custom profile is not specified, the default
profile will be used.

Set up a CLI profile with the following command, substituting `profile_name` with your desired
profile name, and `$AWS_ACCESS_KEY_ID` and `$AWS_SECRET_ACCESS_KEY` environment variables with your
AWS credentials.

`ecs-cli configure profile --profile-name profile_name --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY`

### Cluster Configurations

A cluster configuration is the set of fields that describes an Amazon ECS cluster, including the
name of the cluster and the region. These configurations can then be referenced when you run Amazon
ECS CLI commands using the `--cluster-config` flag; otherwise, the default configuration is used.

Create a cluster configuration with the following command, substituting `region_name` with your
desired AWS region, `cluster_name` with the name of an existing Amazon ECS cluster or a new cluster
to use, and `configuration_name` with the name you'd like to give this configuration.

`ecs-cli configure --cluster cluster_name --region region_name --config-name configuration_name`

#### Defaults

The first Cluster Configuration or ECS Profile that you configure will be set as the default. The
default ECS Profile can be changed using the `ecs-cli configure profile default` command; the
default cluster configuration can be changed using the `ecs-cli configure default` command. Note
that unlike in the AWS CLI, the default ECS Profile does not need to be named "default".

#### Using Credentials from `~/.aws/credentials`, Assuming a Role, and Multi-Factor Authentication

The `--aws-profile` flag and `$AWS_PROFILE` environment variable allow you to reference any named profile in `~/.aws/credentials`.

Here is an example on how to assume a role: [amazon-ecs-cli/blob/master/ecs-cli/modules/config/aws_credentials_example.ini](https://github.com/aws/amazon-ecs-cli/blob/master/ecs-cli/modules/config/aws_credentials_example.ini)

If you are trying to use Multi-Factor Authentication, please see this comment and the associated issue: [#284 (comment)](https://github.com/aws/amazon-ecs-cli/issues/284#issuecomment-336310034).

#### Order of Resolution for credentials

1) ECS CLI Profile Flags
  a) ECS Profile (--ecs-profile)
  b) AWS Profile (--aws-profile)
2) Environment Variables - attempts to fetch the credentials from environment variables:
  a) ECS_PROFILE
  b) AWS_PROFILE
  c) AWS_ACCESS_KEY_ID and AWS_SECRET_ACCESS_KEY, Optional: AWS_SESSION_TOKEN
3) ECS Config - attempts to fetch the credentials from the default ECS Profile
4) Default AWS Profile - attempts to use credentials (aws_access_key_id, aws_secret_access_key) or assume_role (role_arn, source_profile) from AWS profile name
  a) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')
5) EC2 Instance role

#### Order of Resolution for Region

1) ECS CLI Flags
  a) Region Flag --region
  b) Cluster Config Flag (--cluster-config)
2) ECS Config - attempts to fetch the region from the default ECS Profile
3) Environment Variable - attempts to fetch the region from environment variables:
  a) AWS_REGION (OR)
  b) AWS_DEFAULT_REGION
4) AWS Profile - attempts to use region from AWS profile name
  a) AWS_PROFILE environment variable (OR) –aws-
  b) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')

## Using the CLI

### Creating an ECS Cluster
After installing the Amazon ECS CLI and configuring your credentials, you are ready to
create an ECS cluster.


```
$ ecs-cli help up
NAME:
   up - Creates the ECS cluster (if it does not already exist) and the AWS resources required to set up the cluster.

USAGE:
   command up [command options] [arguments...]

OPTIONS:
   --verbose, --debug
   --capability-iam                  Acknowledges that this command may create IAM resources. Required if --instance-role is not specified.
   --keypair value                   [Optional] Specifies the name of an existing Amazon EC2 key pair to enable SSH access to the EC2 instances in your cluster.
   --size value                      [Optional] Specifies the number of instances to launch and register to the cluster. Defaults to 1.
   --azs value                       [Optional] Specifies a comma-separated list of 2 VPC Availability Zones in which to create subnets (these zones must have the available status). This option is recommended if you do not specify a VPC ID with the --vpc option. WARNING: Leaving this option blank can result in failure to launch container instances if an unavailable zone is chosen at random.
   --security-group value            [Optional] Specifies a comma-separated list of existing security groups to associate with your container instances. If you do not specify a security group here, then a new one is created.
   --cidr value                      [Optional] Specifies a CIDR/IP range for the security group to use for container instances in your cluster. This parameter is ignored if an existing security group is specified with the --security-group option. Defaults to 0.0.0.0/0.
   --port value                      [Optional] Specifies a port to open on the security group to use for container instances in your cluster. This parameter is ignored if an existing security group is specified with the --security-group option. Defaults to port 80.
   --subnets value                   [Optional] Specifies a comma-separated list of existing VPC Subnet IDs in which to launch your container instances. This option is required if you specify a VPC with the --vpc option.
   --vpc value                       [Optional] Specifies the ID of an existing VPC in which to launch your container instances. If you specify a VPC ID, you must specify a list of existing subnets in that VPC with the --subnets option. If you do not specify a VPC ID, a new VPC is created with two subnets.
   --instance-type value             [Optional] Specifies the EC2 instance type for your container instances. Defaults to t2.micro.
   --image-id value                  [Optional] Specify the AMI ID for your container instances. Defaults to amazon-ecs-optimized AMI.
   --no-associate-public-ip-address  [Optional] Do not assign public IP addresses to new instances in this VPC. Unless this option is specified, new instances in this VPC receive an automatically assigned public IP address.
   --force, -f                       [Optional] Forces the recreation of any existing resources that match your current configuration. This option is useful for cleaning up stale resources from previous failed attempts.
   --instance-role value             [Optional] Specifies a custom IAM Role for instances in your cluster. Required if --capability-iam is not specified.
   --cluster value, -c value         [Optional] Specifies the ECS cluster name to use. Defaults to the cluster configured using the configure command
   --region value, -r value          [Optional] Specifies the AWS region to use. Defaults to the region configured using the configure command
```

For example, to create an ECS cluster with two Amazon EC2 instances, use the following command:

```
$ ecs-cli up --keypair my-key --capability-iam --size 2
```

It takes a few minutes to create the resources requested by `ecs-cli up`.
To see when the cluster is ready to run tasks, use the AWS CLI to
confirm that the ECS instances are registered:


```
$ aws ecs list-container-instances --cluster your-cluster-name
{
    "containerInstanceArns": [
        "arn:aws:ecs:us-east-1:980116778723:container-instance/6a302e06-0aa6-4bbc-9428-59b17089b887",
        "arn:aws:ecs:us-east-1:980116778723:container-instance/7db3c588-0ef4-49fa-be32-b1e1464f6eb5",
    ]
}

```

**Note:** The default security group created by `ecs-cli up` allows inbound traffic on port 80 by
default. To allow inbound traffic from a different port, specify the port you wish to open with the
`--port` option. To add more ports to the default security group, go to **EC2 Security Groups** in
the AWS Management Console and search for the security group containing “ecs-cli”. Add a rule as
described in the [Adding Rules to a Security Group](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-network-security.html#adding-security-group-rule)
topic.

Alternatively, you may specify one or more existing security group IDs with the `--security-group` option.

### Starting/Running Tasks
After the cluster is created, you can run tasks – groups of containers – on the
ECS cluster. First, author a
[Docker Compose configuration file]( https://docs.docker.com/compose).
You can run the configuration file locally using Docker Compose.

Here is an example Docker Compose configuration file that creates a web page:

```
version: '2'
services:
  web:
    image: amazon/amazon-ecs-sample
    ports:
     - "80:80"
```

To run the configuration file on Amazon ECS, use `ecs-cli compose up`. This creates an ECS task
definition and starts an ECS task. You can see the task that is running with `ecs-cli compose ps`,
for example:

```
$ ecs-cli compose ps
Name                                      State    Ports                     TaskDefinition
fd8d5a69-87c5-46a4-80b6-51918092e600/web  RUNNING  54.209.244.64:80->80/tcp  ecscompose-web:1
```

Navigate your web browser to the task’s IP address to see the sample app running in the ECS cluster.

### Creating a Service
You can also run tasks as services. The ECS service scheduler ensures that the specified number of
tasks are constantly running and reschedules tasks when a task fails (for example, if the underlying
container instance fails for some reason).

```
$ ecs-cli compose --project-name wordpress-test service create

INFO[0000] Using Task definition                         TaskDefinition=ecscompose-wordpress-test:1
INFO[0000] Created an ECS Service                        serviceName=ecscompose-service-wordpress-test taskDefinition=ecscompose-wordpress-test:1

```

You can then start the tasks in your service with the following command:
`$ ecs-cli compose --project-name wordpress-test service start`

It may take a minute for the tasks to start. You can monitor the progress using
the following command:
```
$ ecs-cli compose --project-name wordpress-test service ps
Name                                            State    Ports                      TaskDefinition
34333aa6-e976-4096-991a-0ec4cd5af5bd/wordpress  RUNNING  54.186.138.217:80->80/tcp  ecscompose-wordpress-test:1
34333aa6-e976-4096-991a-0ec4cd5af5bd/mysql      RUNNING                             ecscompose-wordpress-test:1
```

### Using ECS parameters

Since there are certain fields in an ECS task definition that do not correspond to fields in a
Docker Composefile, you can specify those values using the `--ecs-params` flag. Currently, the file
supports the follow schema:

```
version: 1
task_definition:
  ecs_network_mode: string               // Supported string values: none, bridge, host, or awsvpc
  task_role_arn: string
  task_execution_role: string            // Needed to use Cloudwatch Logs or ECR with your ECS tasks
  task_size:                             // Required for running tasks in Fargate mode
    cpu_limit: string
    mem_limit: string
  services:
    <service_name>:
      essential: boolean

run_params:
  network_configuration:
    awsvpc_configuration:
      subnets: array of strings          // These should be in the same VPC and Availability Zone as your instance
      security_groups: array of strings  // These should be in the same VPC as your instance
      assign_public_ip: string           // supported values: ENABLED or DISABLED
```

**Version**
Schema version being used for the ecs-params.yml file. Currently, we only support version 1.

**Task Definition**
Fields listed under `task_definition` correspond to fields that will be included in your ECS Task Definition.

* `ecs_network_mode` corresponds to NetworkMode on an ECS Task Definition (Not to be confused with the network_mode field in Docker Compose). Supported values are none, bridge, host, or awsvpc. If not specified, this will default to bridge mode. If you wish to run tasks with Network Configuration, this field *must* be set to `awsvpc`.

* `task_role_arn` should be the ARN of an IAM role. **NOTE**: If this role does not have the proper permissions/trust relationships on it, the `up` command will fail.

* `services` correspond to the services listed in your docker compose file, with `service_name` matching the name of the container you wish to run. Its fields will be merged into an [ECS Container Definition](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html).
The only field you can specify on it is `essential`. The default value for the essential field is true.

* `task_execution_role` should be the ARN of an IAM role. **NOTE**: This field is required to enable ECS Tasks to be configured with Cloudwatch Logs, or to pull images from ECR for your tasks.

* `task_size` Contains two fields, CPU and Memory. These fields are required for launching tasks in Fargate mode. See [the documentation on ECS Task Definition Parameters](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html) for more information.

**Run Params**
Fields listed under `run_params` are for values needed as options to API calls not related to a Task Definition, such as `compose up` (RunTask) and `compose service up` (CreateService).
Currently, the only parameter supported under `run_params` is `network_configuration`. This is required to run tasks with [Task Networking](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html), as well as in Fargate mode.

* `network_configuration` is required if you specify `ecs_network_mode` as `awsvpc`. It takes one nested parameter, `awsvpc_configuration`, which has three subfields:
  * `subnets`: list of subnet ids used to launch tasks. ***NOTE*** These should be in the same VPC and availability zone as the instances on which you wish to launch your tasks.
  * `security_groups`: list of securtiy-group ids used to launch tasks. ***NOTE*** These should be in the same VPC as the instances on which you wish to launch your tasks.
  * `assign_public_ip`: supported values for this field are either "ENABLED" or "DISABLED". This field is *only* used for tasks launched in Fargate mode. If this field is present in tasks with network configuration launched in EC2 mode, the request will fail.

Example `ecs-params.yml` file:

```
version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: myCustomRole
  services:
    my_service:
      essential: false
```

Example `ecs-params.yml` with network configuration in **EC2** mode:

```
version: 1
task_definition:
  ecs_network_mode: awsvpc
  services:
    my_service:
      essential: false

run_params:
  network_configuration:
    awsvpc_configuration:
      subnets:
        - subnet-feedface
        - subnet-deadbeef
      security_groups:
        - sg-bafff1ed
        - sg-c0ffeefe
```
Example `ecs-params.yml` with network configuration in **FARGATE** mode:

```
version: 1
task_definition:
  ecs_network_mode: awsvpc
  task_execution_role: myFargateRole
  task_size:
    cpu_limit: 512
    mem_limit: 2GB
  services:
    my_service:
      essential: false

run_params:
  network_configuration:
    awsvpc_configuration:
      subnets:
        - subnet-feedface
        - subnet-deadbeef
      security_groups:
        - sg-bafff1ed
        - sg-c0ffeefe
      assign_public_ip: ENABLED
```

You can then start a task by calling:
```
ecs-cli compose --ecs-params my-ecs-params.yml up
```

If you have a file name `ecs-params.yml` in your current directory, `ecs-cli compose` will automatically read it without your having to set the `--ecs-params` flag value explicitly.

```
ecs-cli compose up
```

## Amazon ECS CLI Commands

For a complete list of commands, see the
[Amazon ECS CLI documentation](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI.html).

## Contributing to the CLI
Contributions and feedback are welcome! Proposals and pull requests will be considered and responded to.
For more information, see the [CONTRIBUTING.md](https://github.com/aws/amazon-ecs-cli/blob/master/CONTRIBUTING.md) file.

Amazon Web Services does not currently provide support for modified copies of
this software.

## License

The Amazon ECS CLI is distributed under the [Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0).
See [LICENSE](https://github.com/aws/amazon-ecs-cli/blob/master/LICENSE) and [NOTICE](https://github.com/aws/amazon-ecs-cli/blob/master/NOTICE) for more information.
