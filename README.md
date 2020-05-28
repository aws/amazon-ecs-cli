__✨ The [ECS CLI v2](https://github.com/aws/amazon-ecs-cli-v2) is now in preview: a new way to develop, release and operate your container apps on ECS__

<details>
<summary>Learn more about the ECS CLI V2</summary>

The [ECS CLI v2](https://github.com/aws/amazon-ecs-cli-v2) is a brand new CLI focused on the full developer experience of building, deploying and operating your containerized apps. From helping manage all of your infrastructure, to setting up CD Pipelines, the V2 is here to help. The [ECS CLI v2](https://github.com/aws/amazon-ecs-cli-v2) is still in preview and quite different from V1, but we'd love your feedback! For more info on V2, V1 and how these projects are being developed [check out our V2 proposal](https://github.com/aws/containers-roadmap/issues/513).
</details>


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
CLI. For more information about supported services and to download the AWS CLI, see the [AWS Command
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
		- [Using Route53 Service Discovery](#using-route53-service-discovery)
	- [Viewing Running Tasks](#viewing-running-tasks)
	- [Viewing Container Logs](#viewing-container-logs)
	- [Using FIPS Endpoints](#using-fips-endpoints)
	- [Using Private Registry Authentication](#using-private-registry-authentication)
	- [Checking for Missing Attributes and Debugging Reason Attribute Errors](#checking-for-missing-attributes-and-debugging-reason-attribute-errors)
	- [Tagging Resources](#tagging-resources)
		- [ARN Formats](#arn-formats)
	- [Running Tasks Locally](#running-tasks-locally)
- [Amazon ECS CLI Commands](#amazon-ecs-cli-commands)
- [Contributing to the CLI](#contributing-to-the-cli)
- [License](#license)

#### Security disclosures

If you think you’ve found a potential security issue, please do not post it in the Issues.  Instead, please follow the instructions [here](https://aws.amazon.com/security/vulnerability-reporting/) or email AWS security directly at [aws-security@amazon.com](mailto:aws-security@amazon.com).

## Installing

Download the binary archive for your platform, and install the binary on your `$PATH`.
You can use the provided `md5` hash to verify the integrity of your download.

For information about installing and using the Amazon ECS CLI, see the [ECS Command Line Interface](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI.html).

### Latest version
* Linux:
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-latest](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-latest)
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-latest.md5](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-latest.md5)
* Macintosh:
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-latest](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-latest)
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-latest.md5](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-latest.md5)
* Windows:
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-latest.exe](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-latest.exe)
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-latest.md5](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-latest.md5)

### Download Links for within China

As of v0.6.2 the ECS CLI supports the cn-north-1 region in China. The following links are the exact
same binaries, but they are localized within China to provide a faster download experience.

* Linux:
  * [https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-linux-amd64-latest](https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-linux-amd64-latest)
  * [https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-linux-amd64-latest.md5](https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-linux-amd64-latest.md5)
* Macintosh:
  * [https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-darwin-amd64-latest](https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-darwin-amd64-latest)
  * [https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-darwin-amd64-latest.md5](https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-darwin-amd64-latest.md5)
* Windows:
  * [https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-windows-amd64-latest.exe](https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-windows-amd64-latest.exe)
  * [https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-windows-amd64-latest.md5](https://amazon-ecs-cli.s3.cn-north-1.amazonaws.com.cn/ecs-cli-windows-amd64-latest.md5)

### Download specific version
Using the URLs above, replace `latest` with the desired tag, for example `v1.0.0`. After
downloading, remember to rename the binary file to `ecs-cli`. ***NOTE:*** Windows is only supported starting with version `v1.0.0`.

* Linux:
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-v1.0.0](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-v1.0.0)
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-v1.0.0.md5](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-v1.0.0.md5)
* Macintosh:
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-v1.0.0](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-v1.0.0)
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-v1.0.0.md5](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-v1.0.0.md5)
* Windows:
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-v1.0.0.exe](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-v1.0.0.exe)
  * [https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-v1.0.0.md5](https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-v1.0.0.md5)

### Verifying Signatures

If you wish to verify your ECS CLI download, you can use the PGP Signatures.

#### 1. Install [GnuPG](https://www.gnupg.org/)

###### Linux

Install `gpg` using the package manager on your flavor of linux.

###### Mac

One easy way is to use Homebrew, a package manager for OS X. Install Homebrew using the [instructions on its site](https://brew.sh/).

```bash
brew install gnupg
brew install amazon-ecs-cli
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
curl -o ecs-cli.asc https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-darwin-amd64-latest.asc
```

###### Linux
```
curl -o ecs-cli.asc https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-linux-amd64-latest.asc
```

###### Windows
```
PS C:\> Invoke-WebRequest -OutFile ecs-cli.asc https://amazon-ecs-cli.s3.amazonaws.com/ecs-cli-windows-amd64-latest.exe.asc
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
profile name, and `$AWS_ACCESS_KEY_ID`, `$AWS_SECRET_ACCESS_KEY`, and `AWS_SESSION_TOKEN` environment variables with your
AWS credentials.

`ecs-cli configure profile --profile-name profile_name --access-key $AWS_ACCESS_KEY_ID --secret-key $AWS_SECRET_ACCESS_KEY --session-token AWS_SESSION_TOKEN`

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

#### AMI

You can specify the AMI to use with your EC2 instances using the `--image-id` flag. Alternatively, if you do not specify an image ID, the ECS CLI will use the [recommended Amazon Linux 2 ECS Optimized AMI](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/retrieve-ecs-optimized_AMI.html). By default, the x86 variant of this AMI is used. However, if you specify an instance in the A1 family using `--instance-type`, then the `arm64` version of the ECS Optimized AMI will be used. Note: `arm64` ECS Optimized AMIs are only supported in some regions; please see [Amazon ECS-Optimized Amazon Linux 2 AMI](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/al2ami.html).

#### User Data

For the EC2 launch type, the ECS CLI always creates EC2 instances that include the following User Data:

```
#!/bin/bash
echo ECS_CLUSTER={ clusterName } >> /etc/ecs/ecs.config
```

This user data directs the EC2 instance to join your ECS Cluster. You can optionally include extra user data with `--extra-user-data`; this flag takes a file name as its argument.
The flag can be used multiple times to specify multiple files. Extra user data can be shell scripts or cloud-init directives- see the [EC2 documentation](https://docs.aws.amazon.com/AWSEC2/latest/UserGuide/user-data.html) for more information.
The ECS CLI takes all the User Data, and packs it into a MIME Multipart archive which can be used by cloud-init on the EC2 instance. The ECS CLI even allows existing MIME Multipart archives to be passed in with `--extra-user-data`.
The CLI will unpack the existing archive, and then repack it into the final archive (preserving all header and content type information). Here is an example of specifying extra user data:

```
ecs-cli up \
  --capability-iam \
  --extra-user-data my-shellscript \
  --extra-user-data my-cloud-boot-hook \
  --extra-user-data my-mime-multipart-archive \
  --launch-type EC2
```

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
configuration file locally using Docker Compose. Information about specific compose versions and fields supported by the ecs-cli can be found [here](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/cmd-ecs-cli-compose-parameters.html).

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
    mem_limit: string                    // Values specified without units default to MiB
  pid_mode: string                       // Supported string values: task or host
  ipc_mode: string                       // Supported string values: task, host, or none
  services:
    <service_name>:
      essential: boolean
      repository_credentials:
        credentials_parameter: string
      cpu_shares: integer
      firelens_configuration:
        type: string                     // Supported string values: fluentd or fluentbit
        options: list of strings
      mem_limit: string                  // Values specified without units default to bytes, as in docker run
      mem_reservation: string
      gpu: string
      init_process_enabled: boolean
      healthcheck:
        test: string or list of strings
        interval: string
        timeout: string
        retries: integer
        start_period: string
      logging:
        secret_options:
          - value_from: string
            name: string
      secrets:
        - value_from: string
          name: string
  docker_volumes:
    - name: string
      scope: string                      // Valid values: "shared" | "task"
      autoprovision: boolean             // only valid if scope = "shared"
      driver: string
      driver_opts:
        string: string
      labels:
        string: string
  efs_volumes:
     - name: string
       filesystem_id: string
       root_directory: string
       transit_encryption: string       // Valid values: "ENABLED" | "DISABLED" (default). Required if 
                                        //   IAM is enabled or an access point ID is  
                                        //   specified
       transit_encryption_port: int64   // required if transit_encryption is enabled
       access_point: string
       iam: string                      // Valid values: "ENABLED" | "DISABLED" (default). Enable IAM 
                                        //   authentication for FS access. 
  placement_constraints:
    - type: string                      // Valid values: "memberOf"
      expression: string

run_params:
  network_configuration:
    awsvpc_configuration:
      subnets: array of strings          // These should be in the same VPC and Availability Zone as your instance
      security_groups: list of strings   // These should be in the same VPC as your instance
      assign_public_ip: string           // supported values: ENABLED or DISABLED
  task_placement:
    strategy:
      - type: string                     // Valid values: "spread"|"binpack"|"random"
        field: string                    // Not valid if type is "random"
    constraints:
      - type: string                     // Valid values: "memberOf"|"distinctInstance"
        expression: string               // Not valid if type is "distinctInstance"
  service_discovery:
    container_name: string
    container_port: integer
    private_dns_namespace:
      id: string
      name: string
      vpc: string
      description: string
    public_dns_namespace:
      id: string
      name: string
    service_discovery_service:
      name: string
      description: string
      dns_config:
        type: string
        ttl: integer
      healthcheck_custom_config:
        failure_threshold: integer
```

**Version**
Schema version being used for the ecs-params.yml file. Currently, we only support version 1.

**Task Definition**
Fields listed under `task_definition` correspond to fields that will be included in your ECS Task Definition.

* `ecs_network_mode` corresponds to NetworkMode on an ECS Task Definition (Not to be confused with the network_mode field in Docker Compose). Supported values are none, bridge, host, or awsvpc. If not specified, this will default to bridge mode. If you wish to run tasks with Network Configuration, this field *must* be set to `awsvpc`.

* `task_role_arn` should be the ARN of an IAM role. **NOTE**: If this role does not have the proper permissions/trust relationships on it, the `up` command will fail.

* `services` correspond to the services listed in your docker compose file, with `service_name` matching the name of the container you wish to run. Its fields will be merged into an [ECS Container Definition](http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html).
  * If the [`essential`](https://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/aws-properties-ecs-taskdefinition-containerdefinitions.html#cfn-ecs-taskdefinition-containerdefinition-essential) field is not specified, the value defaults to true.
  * If you are using Docker compose version 3, the `cpu_shares`, `mem_limit`, and `mem_reservation` fields are optional and must be specified in the ECS params file rather than the compose file.
  * In Docker compose version 2, the `cpu_shares`, `mem_limit`, and `mem_reservation` fields can be specified in either the compose or ECS params file. If they are specified in the ECS params file, the values will override values present in the compose file.
  * If you are using a private repository for pulling images, `repository_credentials` allows you to specify an AWS Secrets Manager secret ARN for the name of the secret containing your private repository credentials as a `credential_parameter`.
  * `init_process_enabled` is a [Linux-specific option](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_LinuxParameters.html) that can be be set to run an init process inside the container that forwards signals and reaps processes. This parameter maps to the `--init` option to [docker run](https://docs.docker.com/engine/reference/run/). This parameter requires version 1.25 of the Docker Remote API or greater on your container instance.
  * `firelens_configuration` contains configuration parameters for [Firelens](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_FirelensConfiguration.html).
    * `type` Valid options are fluentbit or fluentd
    * `options` Please see the [AWS docs for Firelens](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_FirelensConfiguration.html)
  * `gpu` is the number of physical GPUs the Amazon ECS container agent will reserve for the container. Maps to the GPU [resource requirement](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_ResourceRequirement.html) field in the task definition. For example: "1", "4", "8", "16".
  * `healthcheck` This parameter maps to `healthcheck` in the [Docker compose file reference](https://docs.docker.com/compose/compose-file/#healthcheck). This field can either be used here in the ECS Params file, or it can be used in Compose File version 3 with the ECS CLI.
    * `test` can also be specified as `command` and must be either a string or a list or strings. If `test` is specified as a list of strings, the first item must be either NONE, CMD, or CMD-SHELL. If test or command is specified as a string, CMD-SHELL will be prepended and ECS will run the command in the container's default shell.
    * `interval`, `timeout`, and `start_period` are specified as durations in a string format. For example: 2.5s, 10s, 1m30s, 2h23m, or 5h34m56s.
  * `secrets` allows you to specify secrets which will be retrieved from SSM Parameter Store. See the [ECS Docs](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/specifying-sensitive-data.html) for more information, including how reference AWS Secrets Managers secrets from SSM Parameter Store.
    * `value_from` is the SSM (or Secrets Manager) Parameter ARN or name (if the parameter is in the same region as your ECS Task).
    * `name` is the name of the environment variable in which the secret will be stored.
  * If you need to inject secrets into your logging configuration, you may set `secret_options` under `logging`. For more information, See the [logging secrets section](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/specifying-sensitive-data.html#secrets-logconfig) of the ECS docs.
    * `value_from` is the SSM (or Secrets Manager) Parameter ARN or name (if the parameter is in the same region as your ECS Task).
    * `name` is the name of the logging option in which the secret will be stored.

* `docker_volumes` allows you to create docker volumes. The name key is required, and `scope`, `autoprovision`, `driver`, `driver_opts` and `labels` correspond with the fields under [dockerVolumeConfiguration](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/docker-volumes.html) in an ECS Task Definition. Volumes defined with the `docker_volumes` key can be referenced in your compose file by name, even if they were not also specified in the compose file.

* `efs_volumes` allows you to mount EFS volumes to your container. The name and EFS filesystem ID are required. EFS volumes can be referenced by name in your compose file like `docker_volumes`. 

* `task_execution_role` should be the ARN of an IAM role. **NOTE**: This field is required to enable ECS Tasks to be configured with Cloudwatch Logs, or to pull images from ECR for your tasks.

* `task_size` Contains two fields, CPU and Memory. These fields are required for launching tasks with Fargate launch type. See [the documentation on ECS Task Definition Parameters](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html) for more information.

* `placement_constraints` allows you to specify a list of constraints on task placement within the task definition. Not supported with the `FARGATE` launch type.

* `pid_mode` allows you to control the process namespace in which your containers run. Valid values are `task` or `host`. See the [ECS documentation](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#task_definition_pidmode) for more information.

* `ipc_mode` allows you to control the IPC resource namespace in which your containers run. Valid values are `task`, `host`, or `none`. See the [ECS documentation](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task_definition_parameters.html#task_definition_ipcmode) for more information.

**Run Params**
Fields listed under `run_params` are for values needed as options to API calls not related to a Task Definition, such as `compose up` (RunTask) and `compose service up` (CreateService).
Currently, the only parameter supported under `run_params` is `network_configuration`. This is required to run tasks with [Task Networking](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-networking.html), as well as with Fargate launch type.

* `network_configuration` is required if you specify `ecs_network_mode` as `awsvpc`. It takes one nested parameter, `awsvpc_configuration`, which has three subfields:
  * `subnets`: list of subnet ids used to launch tasks. ***NOTE*** These should be in the same VPC and availability zone as the instances on which you wish to launch your tasks.
  * `security_groups`: list of securtiy-group ids used to launch tasks. ***NOTE*** These should be in the same VPC as the instances on which you wish to launch your tasks.
  * `assign_public_ip`: supported values for this field are either "ENABLED" or "DISABLED". This field is *only* used for tasks launched with Fargate launch type. If this field is present in tasks with network configuration launched with EC2 launch type, the request will fail.
* `task_placement` is an optional field with `EC2` launch-type only (it is *not* valid for `FARGATE`). It has two subfields:
  * `strategy`: A list of objects, with two keys. Valid keys are `type` and `field`.
    * `type`: Valid values are `random`, `binpack`, or `spread`. If `random` is specified, the `field` key should not be provided.
    * `field`: Valid values depend on the strategy type.
      * For `spread`, valid values are `instanceId`, `host`, or attribute key/value pairs, e.g. `attribute:ecs.instance-type =~ t2.*`
      * For "binpack", valid values are "cpu" or "memory".
  * `constraint`: A list of objects, with two keys. Valid keys are `type` and `expression`.
    * `type`: Valid values are `distinctInstance` and `memberOf`. If `distinctInstance` is specified, the `expression` key should not be provided.
    * `expression`: When `type` is `memberOf`, valid values are key/value pairs for attributes or task groups, e.g. `task:group == databases` or `attribute:color =~ green`.
* `service_discovery` allows the configuration of Service Discovery using Route53 auto naming. For an explanation of these fields, see [Using Route53 Service Discovery](#using-route53-service-discovery).

For more information on task placement, see [Amazon ECS TaskPlacement] (https://docs.aws.amazon.com/AmazonECS/latest/developerguide/task-placement.html).

Example `ecs-params.yml` file:

```
version: 1
task_definition:
  ecs_network_mode: host
  task_role_arn: myCustomRole
  services:
    logging:
      essential: false
    wordpress:
      cpu_shares: 100
      mem_limit: 500m
    mysql:
      cpu_shares: 105
      mem_limit: 500m
      mem_reservation: 450m
  docker_volumes:
    - name: database_volume
      scope: shared
      autoprovision: true
      driver: local
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

Example `ecs-params.yml` with task placement:

```
version: 1
run_params:
  task_placement:
    strategy:
      - field: memory
        type: binpack
      - field: attribute:ecs.availability-zone
        type: spread
      - type: random
    constraints:
      - expression: attribute:ecs.instance-type =~ t2.*
        type: memberOf
      - type: distinctInstance`
```

Example `ecs-params.yml` with EFS volume:

```
version: 1
task_definition:
  task_execution_role: ecsTaskExecutionRole
  ecs_network_mode: awsvpc
  task_size:
    mem_limit: 1.0GB
    cpu_limit: 512
  efs_volumes:
    - name: "myEFSVolume"
      filesystem_id: "fs-fedc8554"
run_params:
  network_configuration:
    awsvpc_configuration:
      subnets:
        - "subnet-0b24acd73f534bb4f"
        - "subnet-0f0e20022e2cccd67"
      security_groups:
        - "sg-0fb24ebc7dd5254b0"
      assign_public_ip: "ENABLED"
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

#### Using Route53 Service Discovery

With the ECS CLI, you can create an ECS Service that uses [Route53 auto naming for service discovery](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/service-discovery.html). Service Discovery requires a Service Discovery Service and a DNS Namespace. Keep in mind that:
* When you enable Service Discovery with the ECS CLI, a new Service Discovery Service is always created using CloudFormation.
* For the DNS Namespace, you have the option of using an existing public or private DNS Namespace, or letting the ECS CLI create a private DNS Namespace for you using CloudFormation.
* Creation of a Public DNS Namespaces is not supported with the ECS CLI.
* Only a single DNS Namespace may be used with Service Discovery.

##### Enabling Service Discovery

###### Specifying Values

The ECS-CLI simplifies the use of Service Discovery by providing default values for most fields, while still allowing maximum configurability. Here are the default values and explanations listed with the ECS Params input schema:

```
version: 1
run_params:
  service_discovery:
    container_name: string            // Required if using SRV records
    container_port: string            // Required if using SRV records
    private_dns_namespace:
      id: string                      // Allows you to specify an existing namespace by ID
      name: string                    // DNS name for private namespace. Either used to specify an existing namespace, or if one does not exist with this name, the ECS CLI will create it
      vpc: string                     // Required if "id" is not specified
      description: string             // Only used if the namespace does not yet exist. Default = "Created by the Amazon ECS CLI"
    public_dns_namespace:
      id: string                      // Specify an existing public namespace by ID
      name: string                    // Or specify an existing public namespace by Name
    service_discovery_service:
      name: string                    // Default = Name of the your ECS Service
      description: string             // Default = "Created by the Amazon ECS CLI"
      dns_config:
        type: string                  // Valid values: A or SRV. SRV is required/the default when using bridge or host network mode. A is the default for the awsvpc network mode.
        ttl: integer                  // Default = 60
      healthcheck_custom_config:
        failure_threshold: integer    // Default = 1
```

###### Simple Workflow

Let's walk through a simple scenario with Service Discovery to see how it works with the ECS CLI. Many of the Service Discovery configuration values can be specified with flags, which take precedence over the ECS Params if both are present. Remember that with the ECS CLI, the Compose Project Name (name of the directory containing your Docker Compose File, unless otherwise specified using the flag) is used as the name for your ECS Service.

First, we create a Service named `backend` and create a Private DNS Namespace in our VPC. Assume that the network mode is `awsvpc`, so the `container_name` and `container_port` values are not needed.

```
$ ecs-cli compose --project-name backend service up --private-dns-namespace tutorial --vpc vpc-04deee8176dce7d7d --enable-service-discovery
INFO[0001] Using ECS task definition                     TaskDefinition="backend:1"
INFO[0002] Waiting for the private DNS namespace to be created...
INFO[0002] Cloudformation stack status                   stackStatus=CREATE_IN_PROGRESS
WARN[0033] Defaulting DNS Type to A because network mode was awsvpc
INFO[0033] Waiting for the Service Discovery Service to be created...
INFO[0034] Cloudformation stack status                   stackStatus=CREATE_IN_PROGRESS
INFO[0065] Created an ECS service                        service=backend taskDefinition="backend:1"
INFO[0066] Updated ECS service successfully              desiredCount=1 serviceName=backend
INFO[0081] (service backend) has started 1 tasks: (task 824b5a76-8f9c-4beb-a64b-6904e320630e).  timestamp="2018-09-12 00:00:26 +0000 UTC"
INFO[0157] Service status                                desiredCount=1 runningCount=1 serviceName=backend
INFO[0157] ECS Service has reached a stable state        desiredCount=1 runningCount=1 serviceName=backend
```

Next, we create another service called `frontend` in the same Private DNS Namespace. Since the Namespace was already created, the ECS CLI knows to use the existing one.

```
$ ecs-cli compose --project-name frontend service up --private-dns-namespace tutorial --vpc vpc-04deee8176dce7d7d --enable-service-discovery
INFO[0001] Using ECS task definition                     TaskDefinition="frontend:1"
INFO[0002] Using existing namespace ns-kvhnzhb5vxplfmls
WARN[0033] Defaulting DNS Type to A because network mode was awsvpc
INFO[0033] Waiting for the Service Discovery Service to be created...
INFO[0034] Cloudformation stack status                   stackStatus=CREATE_IN_PROGRESS
INFO[0065] Created an ECS service                        service=frontend taskDefinition="frontend:1"
INFO[0066] Updated ECS service successfully              desiredCount=1 serviceName=frontend
INFO[0081] (service frontend) has started 1 tasks: (task 824b5a76-8f9c-4beb-a64b-6904e320630e).  timestamp="2018-09-12 00:00:26 +0000 UTC"
INFO[0157] Service status                                desiredCount=1 runningCount=1 serviceName=frontend
INFO[0157] ECS Service has reached a stable state        desiredCount=1 runningCount=1 serviceName=frontend
```

Now, the two Services can find each other in the VPC using DNS. The DNS host name will be the name of the Service Discovery Service plus the name of the DNS Namespace. So the ECS Service `frontend` can be found at `frontend.tutorial`, and `backend` can be found at `backend.tutorial`. Remember that since this is a Private DNS Namespace, these domain names can only be resolved within your VPC.

Now, let's update some of the Service Discovery settings for `frontend`; the only values that can be updated are `DNS TTL` and `Health Check Custom Config Failure Threshold` (the failure threshold for the health check administered by ECS, which determines when unhealthy containers will have their DNS records removed).

```
$ ecs-cli compose --project-name frontend service up --update-service-discovery --dns-type SRV --dns-ttl 120 --healthcheck-custom-config-failure-threshold 2
INFO[0001] Using ECS task definition                     TaskDefinition="frontend:1"
INFO[0001] Updated ECS service successfully              desiredCount=1 serviceName=frontend
INFO[0001] Service status                                desiredCount=1 runningCount=1 serviceName=frontend
INFO[0001] ECS Service has reached a stable state        desiredCount=1 runningCount=1 serviceName=frontend
INFO[0002] Waiting for your Service Discovery resources to be updated...
INFO[0002] Cloudformation stack status                   stackStatus=UPDATE_IN_PROGRESS
```

Next, we delete the services and the Service Discovery resources. When we delete `frontend`, the CLI automatically removes its associated Service Discovery Service.

```
$ ecs-cli compose --project-name frontend service down
INFO[0000] Updated ECS service successfully              desiredCount=0 serviceName=frontend
INFO[0001] Service status                                desiredCount=0 runningCount=1 serviceName=frontend
INFO[0016] Service status                                desiredCount=0 runningCount=0 serviceName=frontend
INFO[0016] (service frontend) has stopped 1 running tasks: (task 824b5a76-8f9c-4beb-a64b-6904e320630e).  timestamp="2018-09-12 00:37:25 +0000 UTC"
INFO[0016] ECS Service has reached a stable state        desiredCount=0 runningCount=0 serviceName=frontend
INFO[0016] Deleted ECS service                           service=frontend
INFO[0016] ECS Service has reached a stable state        desiredCount=0 runningCount=0 serviceName=frontend
INFO[0027] Waiting for your Service Discovery Service resource to be deleted...
INFO[0027] Cloudformation stack status                   stackStatus=DELETE_IN_PROGRESS
```

Finally, we delete `backend` and the Private DNS Namespace which was created with it (the CLI associates the CloudFormation Stack for the Namespace with the ECS Service that it was originally created for, so the two should be deleted together).

```
$ ecs-cli compose --project-name backend service down --delete-namespace
INFO[0000] Updated ECS service successfully              desiredCount=0 serviceName=backend
INFO[0001] Service status                                desiredCount=0 runningCount=1 serviceName=backend
INFO[0016] Service status                                desiredCount=0 runningCount=0 serviceName=backend
INFO[0016] (service backend) has stopped 1 running tasks: (task 824b5a76-8f9c-4beb-a64b-6904e320630e).  timestamp="2018-09-12 00:37:25 +0000 UTC"
INFO[0016] ECS Service has reached a stable state        desiredCount=0 runningCount=0 serviceName=backend
INFO[0016] Deleted ECS service                           service=backend
INFO[0016] ECS Service has reached a stable state        desiredCount=0 runningCount=0 serviceName=backend
INFO[0027] Waiting for your Service Discovery Service resource to be deleted...
INFO[0027] Cloudformation stack status                   stackStatus=DELETE_IN_PROGRESS
INFO[0059] Waiting for your Private DNS Namespace resource to be deleted...
INFO[0059] Cloudformation stack status                   stackStatus=DELETE_IN_PROGRESS
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

You can use the `--desired-status` flag to filter for "STOPPED" or "RUNNING" containers.

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

### Using FIPS Endpoints
The ECS-CLI supports using [FIPS endpoints](https://aws.amazon.com/compliance/fips/) for calls to ECR. To ensure you are accessing ECR using FIPS endpoints, use the `--use-fips` flag on the `push`, `pull`, or `images` command. FIPS endpoints are currently available in us-west-1, us-west-2, us-east-1, us-east-2, and in the [GovCloud partition](https://docs.aws.amazon.com/govcloud-us/latest/ug-west/using-govcloud-endpoints.html).

```
$ ecs-cli push myRepository:latest --use-fips --debug
DEBU[0000] Using FIPS endpoint: https://ecr-fips.us-west-2.amazonaws.com
INFO[0000] Getting AWS account ID...
DEBU[0000] Getting authorization token...
DEBU[0000] Checking file cache                           registry=xxxxxxxxxx123
DEBU[0000] Calling ECR.GetAuthorizationToken             registry=xxxxxxxxxx123
DEBU[0000] Saving credentials to file cache              registry=xxxxxxxxxx123
DEBU[0000] Retrieved authorization token via endpoint: https://xxxxxxxxxxx123.dkr.ecr-fips.us-west-2.amazonaws.com
INFO[0000] Tagging image                                 image=myRepository repository=xxxxxxxxxxx123.dkr.ecr-fips.us-west-2.amazonaws.com/myRepository tag=latest
INFO[0000] Image tagged
DEBU[0000] Check if repository exists                    repository=myRepository
INFO[0000] Pushing image                                 repository=xxxxxxxxxxx123.dkr.ecr-fips.us-west-2.amazonaws.com/myRepository tag=latest
INFO[0002] Image pushed
```

### Using Private Registry Authentication

If you want to use privately hosted container images with ECS, the ECS CLI can store your private registry credentials in AWS Secrets Manager and create an IAM role which ECS can use to access the credentials and private images. This allows you to:

* Store private registry credentials within AWS for use with ECS
* Add the permissions needed to use your registry secrets to a new or existing Task Execution Role
* Automatically add your private registry credentials to your task definition when running a task or service

Using privately hosted images with the ECS CLI is done in two parts:

1) Create new AWS Secrets Manager secrets and an IAM Task Execution Role with `ecs-cli registry-creds up`
2) Run `ecs-cli compose` commands to create and run a task definition that includes the new resources

#### Storing private registry credentials with `ecs-cli registry-creds up`

To get started, first create an input file that contains the name of your registry and the credentials needed to access it:

```
# file name: cred_input.yml
# when using environment variables, only '${VAR_NAME}' format is supported

version: '1'
registry_credentials:
  my-registry.example.com:
    secrets_manager_arn:        # required when using (with no modification) or updating an existing secret
    username: myUserName        # required when creating or updating a new secret
    password: ${MY_PASSWORD}    # required when creating or updating a new secret
    kms_key_id:                 # optional custom KMS Key ID to use to encrypt new secret
    container_names:            # required to match credential resources with docker-compose services
      - web
      - log
```

In this example, we're storing credentials for a registry called `my-registry.example.com` and passing in the password with an environment variable. `container_names` is a list of the `service_names` in your Docker Compose project which need access to images in this registry. If you don't plan to use the output of `registry-creds up` to launch a task or service with `compose`, then you can leave this field empty.

Other options:
* To store credentials for multiple private registries, add additional (up to 10 total) registry names and their required details as separate keys under `registry_credentials`.
  * Existing registry secrets from other regions can be included by specifying their `secrets_manager_arn` and associated `kms_key_id`. Creating or updating secrets must be done from within that region.
* If you want to encrypt the AWS Secrets Manager secret for your registry with a custom KMS Key, then add the ARN, ID or Alias of the Key in the `kms_key_id` field. Otherwise, AWS Secrets Manager will use the default key in your account.
* If you don't want to create or update an IAM Task Execution Role for these secrets, use the `--no-role` flag instead of specifying a role name.
* If you don't want to generate an output file for use with `compose` or for records purposes, use the `--no-output-file` flag.
* If you want the output file to be created in a specific directory on your machine, you can specify it with the `--output-dir <value>` flag. Otherwise, the file will be created in your working directory.

After creating the input file, run the `registry-creds up` command on the file and pass in the name of the new or existing Task Execution Role you want to use for the secrets:

```
$ ecs-cli registry-creds up ./cred_input.yml --role-name myTaskExecutionRole
```

The command will output the names of the resources it creates, including the name of the output file which was generated:

```
$ ecs-cli registry-creds up ./cred_input.yml --role-name myTaskExecutionRole
INFO[0000] Processing credentials for registry my-registry.example.com...
INFO[0000] New credential secret created: arn:aws:secretsmanager:region:aws_account_id:secret:amazon-ecs-cli-setup-my-registry.example.com-VeDqXm
INFO[0000] Creating resources for task execution role myTaskExecutionRole...
INFO[0000] Created new task execution role arn:aws:iam::aws_account_id:role/myTaskExecutionRole
INFO[0000] Created new task execution role policy arn:aws:iam::aws_account_id:policy/amazon-ecs-cli-setup-myTaskExecutionRole-policy-20181023T210805Z
INFO[0000] Attached AWS managed policy arn:aws:iam::aws:policy/service-role/AmazonECSTaskExecutionRolePolicy to role myTaskExecutionRole
INFO[0001] Attached new policy arn:aws:iam::aws_account_id:policy/amazon-ecs-cli-setup-myTaskExecutionRole-policy-20181023T210805Z to role myTaskExecutionRole
INFO[0001] Writing registry credential output to new file C:\Users\myuser\regcreds\regCredTest\ecs-registry-creds_20181023T210805Z.yml
```

The output file `ecs-registry-creds_20181023T210805Z.yml` should like like this:
```
version: "1"
registry_credential_outputs:
  task_execution_role: myTaskExecutionRole
  container_credentials:
    my-registry.example.com:
      credentials_parameter: arn:aws:secretsmanager:region:aws_account_id:secret:amazon-ecs-cli-setup-my-registry.example.com-VeDqXm
      container_names:
      - web
      - log
```

This file contains:
* the name of the IAM Task Execution Role with permissions for the new secrets
* the ARN of the new `credentials_parameter` created for the registry
* the list of containers the new `credentials_parameter` should be used for when running a task or service

We can now use this file with `ecs-cli compose` commands to start a task with images in our private registry.

#### Using private registry credentials when launching tasks or services

Now that we have an output file that identifies which resources we need to use our private registry, the ECS CLI will incorporate them into our Docker Compose project when we run `ecs-cli compose`.

In the same directory (let's call it "privateImageApp"), create a docker-compose.yml file for your application:

```
version: "3"
services:
  web:
    environment:
      - SERVICE_NAME=web
    image: my-registry.example.com/httpd
    ports:
      - "80:80"
  log:
    environment:
      - SERVICE_NAME=log
    image: my-registry.example.com/logging
    logging:
      driver: awslogs
      options:
        awslogs-group: myApps
        awslogs-region: us-west-2
        awslogs-stream-prefix: privateImageApp
```

Now run the command `ecs-cli compose up` to launch a task. The ECS CLI will automatically detect and use the newest `ecs-registry-creds` file within the current directory:

```
$~\privateImageApp> ecs-cli compose up
INFO[0000] Found ecs-registry-creds file C:\Users\myuser\regcreds\regCredTest\ecs-registry-creds_20181023T210805Z.yml
INFO[0000] Using ecs-registry-creds value arn:aws:secretsmanager:region:aws_account_id:secret:amazon-ecs-cli-setup-my-registry.example.com-VeDqXm container name=web option name=credentials_parameter
Using ecs-registry-creds value arn:aws:secretsmanager:region:aws_account_id:secret:amazon-ecs-cli-setup-my-registry.example.com-VeDqXm container name=log option name=credentials_parameter
INFO[0000] Using ecs-registry-creds value myTaskExecutionRole option name=task_execution_role
INFO[0000] Using ECS task definition TaskDefinition="privateImageApp:1"
INFO[0000] Starting container... container=bf35a813-dd76-4fe0-b5a2-c1334c2331f4/web
INFO[0000] Starting container... container=bf35a813-dd76-4fe0-b5a2-c1334c2331f4/log
INFO[0012] Describe ECS container status container=bf35a813-dd76-4fe0-b5a2-c1334c2331f4/web desiredStatus=RUNNING lastStatus=PENDING taskDefinition="privateImageApp:1"
INFO[0013] Describe ECS container status container=bf35a813-dd76-4fe0-b5a2-c1334c2331f4/log desiredStatus=RUNNING lastStatus=PENDING taskDefinition="privateImageApp:1"
INFO[0018] Started container... container=bf35a813-dd76-4fe0-b5a2-c1334c2331f4/web desiredStatus=RUNNING lastStatus=RUNNING taskDefinition="privateImageApp:1"
INFO[0018] Started container... container=bf35a813-dd76-4fe0-b5a2-c1334c2331f4/log desiredStatus=RUNNING lastStatus=RUNNING taskDefinition="privateImageApp:1"
```

 The within your new task definition `privateImageApp:1`, the container definitions for both `web` and `log` should have your "my-registry.example.com" secret as a `credentialsParameter`. The `executionRoleArn` field will be the role we created in the previous step, "myTaskExecutionRole".

 Other options:
 * to use an ecs-registry-creds output file from outside the current directory, you can specify it in with the `--registry-creds <value>` flag

 For more information about using private registries with ECS, see [Private Registry Authentication for Tasks](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/private-auth.html).

### Checking for Missing Attributes and Debugging Reason Attribute Errors

Sometimes, when you try to Run a Task, the API will return the error message `"Reasons : ["ATTRIBUTE"]"`. This occurs because your container instances are missing an attribute required by your Task Definition. You can debug these failures using the `ecs-cli check-attributes` command.

Here's an example of the command in action:

```
$ ecs-cli check-attributes --container-instances 28c5abd2-360e-41a0-81d8-0afca2d08d9b,45510138-f24f-47c6-a418-71c46dd51f88,ae66e18e-1d46-47ff-81c5-647f0f1426ce,dffe7f91-8faa-4e00-983b-c58fd279cf6d --cluster practice-cluster --region us-east-2 --task-def fluentd-kinesis
Container Instance                    Missing Attributes
dffe7f91-8faa-4e00-983b-c58fd279cf6d  None
28c5abd2-360e-41a0-81d8-0afca2d08d9b  com.amazonaws.ecs.capability.logging-driver.fluentd
45510138-f24f-47c6-a418-71c46dd51f88  None
ae66e18e-1d46-47ff-81c5-647f0f1426ce  com.amazonaws.ecs.capability.logging-driver.fluentd
```

The command outputs a table of container instances and which attributes they are missing. In this case, the Task Definition requires the Fluentd log driver, but 2 container instances lack support for it.

### Tagging Resources

ECS CLI Commmands support a `--tags` flag which allows you to specify AWS Resource Tags in the format `key=value,key2=value2,key3=value3`. Resource tags can be used for cost allocation, automation, access control, and more. See [AWS Tagging Strategies](https://aws.amazon.com/answers/account-management/aws-tagging-strategies/) for a discussion of use cases.

#### ARN Formats

ECS has released [new longer ARN formats](https://aws.amazon.com/blogs/compute/migrating-your-amazon-ecs-deployment-to-the-new-arn-and-resource-id-format-2/). ***You must opt in to these new formats in order to tag Tasks, Services, and Container instances.*** We strongly recommend opting-in all IAM Identities in your account. You can use the [PutAccountSettingDefault](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_PutAccountSettingDefault.html) API to opt-in to the new format for all IAM Identities in your account.


#### ecs-cli up command

 The ECS Cluster, and CloudFormation template with EC2 resources can be tagged. In addition, the ECS CLI will add tags to the following resources which are created by the CloudFormation template:
 * VPC
 * Subnets
 * Internet Gateway
 * Route Tables
 * Security Group
 * Autoscaling Group
 * ECS Container Instances (only if opted-in to [Container Instance Long ARN format](https://aws.amazon.com/blogs/compute/migrating-your-amazon-ecs-deployment-to-the-new-arn-and-resource-id-format-2/))

 For the autoscaling group, the ECS CLI will add a `Name` tag whose value will be `ECS Instance - <CloudFormation stack name>`, which will be propagated to your EC2 instances. You can override this behavior by specifying your own `Name` tag.

#### ecs-cli compose create/up

Resource tags specified with `--tags` will be added to your Tasks and Task Definitions. In addition, [ECS Managed Tags](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-using-tags.html) are enabled by default for all tasks launched by the ECS CLI (if you are opted-in the the new Task Long ARN Format). ECS will automatically add a `aws:ecs:clusterName` tag to each of your tasks. You can disable this feature using `--disable-ecs-managed-tags`.

#### ecs-cli compose service create/up

Resource tags specified with `--tags` will be added to your Service and Task Definitions. In addition, all Services created by the ECS CLI have [`propagateTags`](https://docs.aws.amazon.com/AmazonECS/latest/APIReference/API_CreateService.html#ECS-CreateService-request-propagateTags) set to `TASK_DEFINITION` which means that tags from the Task Definition will propagate to the tasks in the Service. If you add new tags, the ECS CLI will register a new Task Definition and these tags will be propagated by ECS to your tasks.

Similar to `compose up/create`, [ECS Managed Tags](https://docs.aws.amazon.com/AmazonECS/latest/developerguide/ecs-using-tags.html) are enabled by default for all Services launched by the ECS CLI (if you are opted-in the the new Task Long ARN Format). ECS will automatically add `aws:ecs:clusterName` and `aws:ecs:serviceName` tags to each of the tasks launched by your service. You can disable this feature using `--disable-ecs-managed-tags`.

#### ecs-cli push

Resource tags specified with `--tags` will be added to your ECR repository.

#### ecs-cli registry-creds up

Resource tags specified with `--tags` will be added to new IAM Roles and new or existing AWS Secrets Manager Secrets. (Existing IAM Roles cannot be tagged.)

### Running Tasks Locally
The ECS CLI supports creating, running, inspecting and stopping tasks defined by an ECS Task Definition through its `local` subcommands. You can run an ECS Task Definition specified in a local JSON file or pulled from a registered ECS Task Definition.

#### ecs-cli local create
If you want to convert an ECS Task Definition to a Docker Compose file, you can run:

```
$ ecs-cli local create
```
Without arguments, this will try to read an ECS Task Definition from local a file named `task-definition.json` located in the current directory and generate both a compose file, by default named `docker-compose.ecs-local.yml`, as well as a compose override file, by default named `docker-compose.ecs-local.override.yml`. This command is equivalent to a dry-run of `local up`.
**NOTE** Using these Compose files as input to `ecs-cli compose` subcommands may not translate back to the same ECS Task Definition used as input to `local create`.

To run an ECS Task Definition specified in a different file, you can use the `--task-def-file` or `-f` flag with the name of the file.
To run an ECS Task Definition already registered with ECS, you can use the `--task-def-remote` or `-t` flag with the ARN or family name of the Task Definition.
You can also specify a different output file using the `--output` or `-o` flag.
To skip the overwrite confirmation prompt, use the `--force` flag.


#### ecs-cli local up
To run an ECS Task Definition locally, you can run:

```
$ ecs-cli local up
```

This command takes the same flags as `local create`. You can also specify compose override files using the `--override` flag.

This command will also create the local end [Amazon ECS Local Endpoints Container](https://github.com/awslabs/amazon-ecs-local-container-endpoints) and the network, `ecs-local-network` that your containers will be run in.


#### ecs-cli local ps
Once you have your task running locally, the basic command to list your task's containers is:
 ```
$ ecs-cli local ps
```
This will search for containers created from the `./task-definition.json` file (to see all available options, run `ecs-cli local ps --help`).

For example, if you'd like to list containers created from a specific task definition file, use the following command:
```
$ ecs-cli local ps -f ./app-task-definition.json
CONTAINER ID        IMAGE               STATUS              PORTS               NAMES                 TASKDEFINITION
84ff8e68e613        nginx               Up 15 seconds                           /local-cmds_nginx_1   /path/to/app-task-definition.json
```

#### ecs-cli local down
If you want to stop and remove a task's containers, you can run:
```
$ ecs-cli local down
```
This will stop and remove all the containers started from the `./task-definition.json` file  (to see all available options, run `ecs-cli local down --help`).

For example, you can stop and remove all tasks running locally using the `--all` flag:
```
$ ecs-cli local down --all
INFO[0000] Searching for all running containers
INFO[0000] Stop and remove 1 container(s)
INFO[0000] Stopped container with id 84ff8e68e613
INFO[0000] Removed container with id 84ff8e68e613
INFO[0000] The network ecs-local-network has no more running tasks
INFO[0001] Stopped container with name amazon-ecs-local-container-endpoints
INFO[0001] Removed container with name amazon-ecs-local-container-endpoints
INFO[0001] Removed network with name ecs-local-network
```

If you have no more tasks running, then this command will also stop and remove the [Amazon ECS Local Container Endpoints](https://github.com/awslabs/amazon-ecs-local-container-endpoints)
and finally remove the `ecs-local-network` as well.

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
