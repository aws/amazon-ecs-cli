# Amazon ECS CLI

The Amazon ECS Command Line Interface (CLI) is a command line interface for Amazon
EC2 Container Service (Amazon ECS) that provides high-level commands to simplify
creating, updating, and monitoring clusters and tasks from a local development
environment. The Amazon ECS CLI supports
[Docker Compose](https://docs.docker.com/compose/), a popular open-source tool
for defining and running multi-container applications. Use the CLI as part
of your everyday development and testing cycle as an alternative to the AWS
Management Console.

For more information about Amazon ECS, see the
[Amazon ECS Developer Guide](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/Welcome.html).
For information about installing and using the Amazon ECS CLI, see the
[ECS Command Line Interface](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI.html).

The AWS Command Line Interface (AWS CLI) is a unified client for AWS services
that provides commands for all public API operations. These commands are lower
level than those provided by the Amazon ECS CLI. For more information about supported
services and to download the AWS CLI, see the
[AWS Command Line Interface](http://aws.amazon.com/cli/) product detail page.

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
  * (Not yet implemented)

#### Download Links for within China

As of v0.6.2 the ECS CLI supports the cn-north-1 region in China. The following links are the exact same binaries, but they are localized within China to provide a faster download experience.

* Linux:
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-linux-amd64-latest](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-linux-amd64-latest)
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-linux-amd64-latest.md5](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-linux-amd64-latest.md5)
* Macintosh:
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-darwin-amd64-latest](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-darwin-amd64-latest)
  * [https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-darwin-amd64-latest.md5](https://s3.cn-north-1.amazonaws.com.cn/amazon-ecs-cli/ecs-cli-darwin-amd64-latest.md5)
  * Windows:
    * (Not yet implemented)

### Download specific version
Using the URLs above, replace `latest` with the desired tag, for example `v0.4.1`. After downloading, remember to rename the binary file to `ecs-cli`.

* Linux:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v0.4.1](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v0.4.1)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v0.4.1.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v0.4.1.md5)
* Macintosh:
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v0.4.1](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v0.4.1)
  * [https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v0.4.1.md5](https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v0.4.1.md5)
* Windows:
  * (Not yet implemented)

## Configuring the CLI

Before using the CLI, you need to configure your AWS credentials, the AWS
region in which to create your cluster, and the name of the ECS cluster to use
with the `ecs-cli configure` command. These settings are stored in
`~/.ecs/config`. You can use any existing AWS named profiles in
`~/.aws/credentials` for your credentials with the `--profile` option.

```
$ ecs-cli help configure
NAME:
   configure - Configures your AWS credentials, the AWS region to use, and the ECS cluster name to use with the Amazon ECS CLI. The resulting configuration is stored in the ~/.ecs/config file.

USAGE:
   command configure [command options] [arguments...]

OPTIONS:
   --region, -r 					Specifies the AWS region to use. If the AWS_REGION environment variable is set when ecs-cli configure is run, then the AWS region is set to the value of that environment variable. [$AWS_REGION]
   --access-key 					Specifies the AWS access key to use. If the AWS_ACCESS_KEY_ID environment variable is set when ecs-cli configure is run, then the AWS access key ID is set to the value of that environment variable. [$AWS_ACCESS_KEY_ID]
   --secret-key 					Specifies the AWS secret key to use. If the AWS_SECRET_ACCESS_KEY environment variable is set when ecs-cli configure is run, then the AWS secret access key is set to the value of that environment variable. [$AWS_SECRET_ACCESS_KEY]
   --profile, -p 					Specifies your AWS credentials with an existing named profile from ~/.aws/credentials. If the AWS_PROFILE environment variable is set when ecs-cli configure is run, then the AWS named profile is set to the value of that environment variable. [$AWS_PROFILE]
   --cluster, -c 					Specifies the ECS cluster name to use. If the cluster does not exist, it is created when you try to add resources to it with the ecs-cli up command.
   --compose-project-name-prefix "ecscompose-"		[Optional] Specifies the prefix added to an ECS task definition created from a compose file. Format <prefix><project-name>.
   --compose-service-name-prefix "ecscompose-service-"	[Optional] Specifies the prefix added to an ECS service created from a compose file. Format <prefix><project-name>.
   --cfn-stack-name-prefix "amazon-ecs-cli-setup-"	[Optional] Specifies the prefix added to the AWS CloudFormation stack created on ecs-cli up. Format <prefix><cluster-name>.
```

## Using the CLI
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
   --keypair value                   Specifies the name of an existing Amazon EC2 key pair to enable SSH access to the EC2 instances in your cluster.
   --capability-iam                  Acknowledges that this command may create IAM resources. Required if --instance-role is not specified.
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

**Note:** The default security group created by `ecs-cli up` allows inbound
traffic on port 80 by default. To allow inbound traffic from a different port,
specify the port you wish to open with the `--port` option. To add more ports
to the default security group, go to **EC2 Security Groups** in the AWS Management
Console and search for the security group containing “ecs-cli”. Add a rule as
described in the
[Adding Rules to a Security Group](http://docs.aws.amazon.com/AWSEC2/latest/UserGuide/using-network-security.html#adding-security-group-rule) topic.
Alternatively, you may specify one or more existing security group IDs with the
`--security-group` option.

After the cluster is created, you can run tasks – groups of containers – on the
ECS cluster. First, author a
[Docker Compose configuration file]( https://docs.docker.com/compose).
You can run the configuration file locally using Docker Compose. Here is an
example Docker Compose configuration file that creates a web page:

```
version: '2'
services:
  web:
    image: amazon/amazon-ecs-sample
    ports:
     - "80:80"
```

To run the configuration file on Amazon ECS, use `ecs-cli compose up`. This
creates an ECS task definition and starts an ECS task. You can see the task
that is running with `ecs-cli compose ps`, for example:

```
$ ecs-cli compose ps
Name                                      State    Ports                     TaskDefinition
fd8d5a69-87c5-46a4-80b6-51918092e600/web  RUNNING  54.209.244.64:80->80/tcp  ecscompose-web:1
```

Navigate your web browser to the task’s IP address to see the sample app
running in the ECS cluster.

You can also run tasks as services. The ECS service scheduler ensures that the
specified number of tasks are constantly running and reschedules tasks when a
task fails (for example, if the underlying container instance fails for some
reason).

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

## Amazon ECS CLI Commands

For a complete list of commands, see the
[Amazon ECS CLI documentation](http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI.html).

## Contributing to the CLI
Contributions and feedback are welcome! Proposals and pull requests will be
considered and responded to. For more information, see
[CONTRIBUTING.md](https://github.com/aws/amazon-ecs-cli/blob/master/CONTRIBUTING.md)
file.

Amazon Web Services does not currently provide support for modified copies of
this software.

## License

The Amazon ECS CLI is distributed under the
[Apache License, Version 2.0](http://www.apache.org/licenses/LICENSE-2.0). See [LICENSE](https://github.com/aws/amazon-ecs-cli/blob/master/LICENSE) and [NOTICE](https://github.com/aws/amazon-ecs-cli/blob/master/NOTICE) for more information.
