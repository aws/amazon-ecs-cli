# Amazon ECS CLI

The Amazon ECS Command Line Interface (CLI) is a command line tool for Amazon Elastic Container
Service (Amazon ECS) that provides high-level commands to simplify creating, updating, and
monitoring clusters and tasks from a local development environment. The Amazon ECS CLI supports
[Docker Compose](https://docs.docker.com/compose/), a popular open-source tool for defining and
running multi-container applications. Use the CLI as part of your everyday development and testing
cycle as an alternative to the AWS Management Console or the AWS CLI.

For more information about Amazon ECS, see the [Amazon ECS Developer
Guide](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/Welcome.html).

The AWS Command Line Interface (AWS CLI) is a unified client for AWS services that provides commands
for all public API operations. These commands are lower level than those provided by the Amazon ECS
CLI. For more information about supported services and to download the AWS CLI, see the [AWS Command
Line Interface](http://aws.amazon.com/cli/) product detail page.

- [Installing](#installing)
	- [Latest version](#latest-version)
	- [Download Links for within China](#download-links-for-within-china)
	- [Download specific version](#download-specific-version)
	- [Verifying Signatures](#verifying-signatures)
- [Configuring the CLI](#configuring-the-cli)
	- [ECS Profiles](#ecs-profiles)
	- [Cluster Configurations](#cluster-configurations)
	- [Configuring Defaults](#configuring-defaults)
- [Using the CLI](#using-the-cli)
	- [Creating an ECS Cluster](#creating-an-ecs-cluster)
		- [Creating a Fargate cluster](#creating-a-fargate-cluster)
	- [Starting/Running Tasks](#startingrunning-tasks)
	- [Creating a Service](#creating-a-service)
	- [Using ECS parameters](#using-ecs-parameters)
		- [Launching an AWS Fargate task](#launching-an-aws-fargate-task)
	- [Viewing Running Tasks](#viewing-running-tasks)
	- [Viewing Container Logs](#viewing-container-logs)
- [Amazon ECS CLI Commands](#amazon-ecs-cli-commands)
- [Contributing to the CLI](#contributing-to-the-cli)
- [License](#license)

## Installing

Download the binary archive for your platform, and install the binary on your `$PATH`.
You can use the provided `md5` hash to verify the integrity of your download.

For information about installing and using the Amazon ECS CLI, see the [ECS Command Line Interface](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI.html).

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

### Download Links for within China

As of v0.6.2 the ECS CLI supports the cn-north-1 region in China. The following links are the exact
same binaries, but they are localized within China to provide a faster download experience.

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
Using the URLs above, replace `latest` with the desired tag, for example `v1.0.0`. After
downloading, remember to rename the binary file to `ecs-cli`.
'''NOTE''': Windows is only supported starting with version `v1.0.0`.

* Linux:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v1.0.0](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v1.0.0)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v1.0.0.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v1.0.0.md5)
* Macintosh:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v1.0.0](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v1.0.0)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v1.0.0.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v1.0.0.md5)
* Windows:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-v1.0.0.exe](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-v1.0.0.exe)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-v1.0.0.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-v1.0.0.md5)

### Verifying Signatures

If you wish to verify your ECS CLI download, you can use the PGP Signatures.

#### 1. Install [GnuPG](https://www.gnupg.org/)

###### Linux

Install `gpg` using the package manager on your flavor of linux.

###### Mac

One easy way is to use Homebrew, a package manager for OS X. Install Homebrew using the [instructions on its site](https://brew.sh/).

```
brew install gnupg
```

###### Windows

Go to the GnuPG [download page](https://gnupg.org/download/) and download the simple installer for Windows. Use the installer to install the GPG tool.

#### 2. Import the Amazon ECS PGP Public Key

You can find the Public Key in our GitHub Repo, in the file [amazon-ecs-public-key.gpg](amazon-ecs-public-key.gpg).

```
gpg --import amazon-ecs-public-key.gpg
```

Key Metadata:

- Key ID: 0x2D51784F
- Type: RSA
- Size: 4096/4096
- Expires: Never
- User ID: Amazon ECS <ecs-security@amazon.com>
- Key fingerprint: F34C 3DDA E729 26B0 79BE AEC6 BCE9 D9A4 2D51 784F

#### 4. Downloading Signatures

ECS CLI signatures are ascii armored detached PGP signatures stored in files with the extension ".asc". The signatures file will have the same name as its corresponding executable with ".asc" appended. In the

###### Mac
```
curl -o ecs-cli.asc https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-latest.asc
```

###### Linux
```
curl -o ecs-cli.asc https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-latest.asc
```

###### Windows
```
PS C:\> Invoke-WebRequest -OutFile ecs-cli.asc https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-windows-amd64-latest.exe.asc
```
#### 4. Verifying a Signature

Assuming you installed the ECS CLI in the recommended location for your platform:

###### Mac and Linux
```
gpg --verify ecs-cli.asc /usr/local/bin/ecs-cli
```
###### Windows
```
gpg --verify ecs-cli.asc C:\Program Files\Amazon\ECSCLI\ecs-cli.exe
```

Expected output:

```
gpg: Signature made Tue Apr  3 13:29:30 2018 PDT
gpg:                using RSA key DE3CBD61ADAF8B8E
gpg: Good signature from "Amazon ECS <ecs-security@amazon.com>" [unknown]
gpg: WARNING: This key is not certified with a trusted signature!
gpg:          There is no indication that the signature belongs to the owner.
Primary key fingerprint: F34C 3DDA E729 26B0 79BE  AEC6 BCE9 D9A4 2D51 784F
     Subkey fingerprint: EB3D F841 E2C9 212A 2BD4  2232 DE3C BD61 ADAF 8B8E
```

The warning in the output is expected and is not problematic; it occurs because there is not a chain of trust between your personal PGP key (if you have one) and the Amazon ECS PGP key. For more information, learn about the [Web of trust](https://en.wikipedia.org/wiki/Web_of_trust).


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

You can also optionally add `--default-launch-type` to your cluster configuration. This value will
be used as the launch type for tasks run in this cluster (see: [Launching an AWS Fargate
Task](#launching-an-aws-fargate-task)) , and will also be used to determine which resources to
create when you bring up a cluster (see: [Creating a Fargate Cluster](#creating-a-fargate-cluster)).
Valid values for this field are EC2 or FARGATE. If not specified, ECS will default to EC2 launch
type.

### Configuring Defaults

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
4)  AWS Profile - attempts to use region from AWS profile name
   a) AWS_PROFILE environment variable (OR) –aws-
   b) AWS_DEFAULT_PROFILE environment variable (defaults to 'default')


For more information, see [ECS CLI Configuration](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI_Configuration.html).

## Using the CLI

ECS now offers two different launch types for tasks and services: EC2 and FARGATE. With the FARGATE
launch type, customers no longer have to manage their own container-instances.

In the ECS-CLI, you can specify either launch type when you bring up a cluster using the
`--launch-type` flag (see: [Creating an ECS Cluster](#creating-an-ecs-cluster)). You can also
configure your cluster to use a particular launch type with the `--default-launch-type` flag (see:
[Cluster Configurations](#cluster-configurations)).

You can also specify which launch type to use for a task or service in `compose up` or `compose
service up`, regardless of which launch type is configured for your cluster (see: [Starting/Running
Tasks](#startingrunning-tasks)).

### Creating an ECS Cluster
After installing the Amazon ECS CLI and configuring your credentials, you are ready to create an ECS cluster. The basic command for creating a cluster is:
```
ecs-cli up
```

(To see all available options, run `ecs-cli up --help`)

For example, to create an ECS cluster with two Amazon EC2 instances using the EC2 launch type, use
the following command:

```
$ ecs-cli up --keypair my-key --capability-iam --size 2
```

It takes a few minutes to create the resources requested by `ecs-cli up`.  To see when the cluster
is ready to run tasks, use the AWS CLI to confirm that the ECS instances are registered:

```
$ aws ecs list-container-instances --cluster your-cluster-name
{
    "containerInstanceArns": [
        "arn:aws:ecs:us-east-1:980116778723:container-instance/6a302e06-0aa6-4bbc-9428-59b17089b887",
        "arn:aws:ecs:us-east-1:980116778723:container-instance/7db3c588-0ef4-49fa-be32-b1e1464f6eb5",
    ]
}

```
In addition to EC2 Instances, other resources created by default include:
* Autoscaling Group
* Autoscaling Launch Configuration
* EC2 VPC
* EC2 Internet Gateway
* EC2 VPC Gateway Attachment
* EC2 Route Table
* EC2 Route
* 2 Public EC2 Subnets
* 2 EC2 SubnetRouteTableAssocitaions
* EC2 Security Group

You can provide your own resources (such as subnets, VPC, or security groups) via their flag options.

**Note:** The default security group created by `ecs-cli up` allows inbound traffic on port 80 by
default. To allow inbound traffic from a different port, specify the port you wish to open with the
`--port` option. To add more ports to the default security group, go to **EC2 Security Groups** in
the AWS Management Console and search for the security group containing “ecs-cli”. Add a rule as
described in the [Adding Rules to a Security Group](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-network-security.html#adding-security-group-rule)
topic.

Alternatively, you may specify one or more existing security group IDs with the `--security-group` option.

You can also create an empty ECS cluster by using the `--empty` or `--e` flag:

```
ecs-cli up --cluster myCluster --empty
```

This is equivalent to the [create-cluster command](https://docs.aws.amazon.com/cli/latest/reference/ecs/create-cluster.html), and will not create a CloudFormation stack associated with your cluster.

#### Creating a Fargate cluster

```
$ ecs-cli up --launch-type FARGATE
```

This will create an ECS Cluster without container instances. By default, this will create the
following resources:

* EC2 VPC
* EC2 Internet Gateway
* EC2 VPC Gateway Attachment
* EC2 Route Table
* EC2 Route
* 2 Public EC2 Subnets
* 2 EC2 SubnetRouteTableAssocitaions

The subnet and VPC ids will be printed to the terminal once the creation is complete. You can then
use the subnet IDs in your ECS Params file to launch Fargate tasks.

For more information on using AWS Fargate, see the [ECS CLI Fargate tutorial](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI_tutorial_fargate.html).

### Starting/Running Tasks
After the cluster is created, you can run tasks – groups of containers – on the ECS cluster. First,
author a [Docker Compose configuration file](https://docs.docker.com/compose).  You can run the
configuration file locally using Docker Compose.

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
fd8d5a69-87c5-46a4-80b6-51918092e600/web  RUNNING  54.209.244.64:80->80/tcp  web:1
```

Navigate your web browser to the task’s IP address to see the sample app running in the ECS cluster.

### Creating a Service
You can also run tasks as services. The ECS service scheduler ensures that the specified number of
tasks are constantly running and reschedules tasks when a task fails (for example, if the underlying
container instance fails for some reason).

```
$ ecs-cli compose --project-name wordpress-test service create

INFO[0000] Using Task definition                         TaskDefinition=wordpress-test:1
INFO[0000] Created an ECS Service                        serviceName=wordpress-test taskDefinition=wordpress-test:1

```

You can then start the tasks in your service with the following command:
`$ ecs-cli compose --project-name wordpress-test service start`

It may take a minute for the tasks to start. You can monitor the progress using
the following command:
```
$ ecs-cli compose --project-name wordpress-test service ps
Name                                            State    Ports                      TaskDefinition
34333aa6-e976-4096-991a-0ec4cd5af5bd/wordpress  RUNNING  54.186.138.217:80->80/tcp  wordpress-test:1
34333aa6-e976-4096-991a-0ec4cd5af5bd/mysql      RUNNING                             wordpress-test:1
```

See the `$ ecs-cli compose service` [documentation page](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cmd-ecs-cli-compose-service.html) for more information about available service options, including load balancing.

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
  task_size:                             // Required for running tasks with Fargate launch type
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

* `task_size` Contains two fields, CPU and Memory. These fields are required for launching tasks with Fargate launch type. See [the documentation on ECS Task Definition Parameters](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html) for more information.

**Run Params**
Fields listed under `run_params` are for values needed as options to API calls not related to a Task Definition, such as `compose up` (RunTask) and `compose service up` (CreateService).
Currently, the only parameter supported under `run_params` is `network_configuration`. This is required to run tasks with [Task Networking](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html), as well as with Fargate launch type.

* `network_configuration` is required if you specify `ecs_network_mode` as `awsvpc`. It takes one nested parameter, `awsvpc_configuration`, which has three subfields:
  * `subnets`: list of subnet ids used to launch tasks. ***NOTE*** These should be in the same VPC and availability zone as the instances on which you wish to launch your tasks.
  * `security_groups`: list of securtiy-group ids used to launch tasks. ***NOTE*** These should be in the same VPC as the instances on which you wish to launch your tasks.
  * `assign_public_ip`: supported values for this field are either "ENABLED" or "DISABLED". This field is *only* used for tasks launched with Fargate launch type. If this field is present in tasks with network configuration launched with EC2 launch type, the request will fail.

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

Example `ecs-params.yml` with network configuration with **EC2** launch type:

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
Example `ecs-params.yml` with network configuration with **FARGATE** launch type:

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

#### Launching an AWS Fargate task

With network configuration specified in your ecs-params.yml file, you can now launch a task with
launch type FARGATE:

```
ecs-cli compose --ecs-params my-ecs-params.yml up --launch-type FARGATE
```

or

```
ecs-cli compose --ecs-params my-ecs-params.yml service up --launch-type FARGATE
```

### Viewing Running Tasks

The PS commands allow you to see running and recently stopped tasks. To see the Tasks running in your cluster:

```
$ ecs-cli ps
Name                                            State    Ports                     TaskDefinition
37e873f6-37b4-42a7-af47-eac7275c6152/web        RUNNING  10.0.1.27:8080->8080/tcp  TaskNetworking:2
37e873f6-37b4-42a7-af47-eac7275c6152/lb         RUNNING  10.0.1.27:80->80/tcp      TaskNetworking:2
37e873f6-37b4-42a7-af47-eac7275c6152/redis      RUNNING                            TaskNetworking:2
40bedf31-d707-446e-affc-766eac4cfb85/mysql      RUNNING                            fargate:1
40bedf31-d707-446e-affc-766eac4cfb85/wordpress  RUNNING  54.16.93.6:80->80/tcp     fargate:1
```

The IP address displayed by the ECS CLI depends on how your cluster is configured and which launch-type is used. If you are running tasks with launch type EC2 without task networking, then the IP address shown will be the public IP of the EC2 instance running your task. If no public IP was assigned, the instance's private IP will be displayed.

For tasks that use Task Networking with EC2 launch type, the ECS CLI will only show the private IP address of the ENI attached to the task.

For Fargate tasks, the ECS CLI will return the public IP assigned to the ENI attached to the Fargate task. The ENI for your Fargate task will be assigned a public IP if `assign_public_ip: ENABLED` is present in your ECS Params file. If the ENI lacks a public IP, then its private IP is shown.

### Viewing Container Logs

View the CloudWatch Logs for a given task and container:

`ecs-cli logs --task-id 4c2df707-a160-475e-9c16-15dfb9df01cc --container-name mysql`

For Fargate tasks, it is recommended that you send your container logs to CloudWatch. *Note: For Fargate tasks you must specify a Task Execution IAM Role in your ECS Params file in order to use CloudWatch Logs.* You can specify the `awslogs` driver and logging options in your compose file like this:

```
services:
  <My Service>:
    logging:
      driver: awslogs
      options:
        awslogs-group: <Log Group Name>
        awslogs-region: <Log Region>
        awslogs-stream-prefix: <Prefix Name>
```

The log stream prefix is technically optional; however, it is highly recommended that you specify it. If you do specify it, then you can use the `ecs-cli logs` command. The Logs command allows you to retrieve the Logs for a task. There are many options for the logs command:

```
OPTIONS:
--task-id value            Print the logs for this ECS Task.
--task-def value           [Optional] Specifies the name or full Amazon Resource Name (ARN) of the ECS Task Definition associated with the Task ID. This is only needed if the Task is using an inactive Task Definition.
--follow                   [Optional] Specifies if the logs should be streamed.
--filter-pattern value     [Optional] Substring to search for within the logs.
--container-name value     [Optional] Prints the logs for the given container. Required if containers in the Task use different log groups
--since value              [Optional] Returns logs newer than a relative duration in minutes. Cannot be used with --start-time (default: 0)
--start-time value         [Optional] Returns logs after a specific date (format: RFC 3339. Example: 2006-01-02T15:04:05+07:00). Cannot be used with --since flag
--end-time value           [Optional] Returns logs before a specific date (format: RFC 3339. Example: 2006-01-02T15:04:05+07:00). Cannot be used with --follow
--timestamps, -t           [Optional] Shows timestamps on each line in the log output.
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
