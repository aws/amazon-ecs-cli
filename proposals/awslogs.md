# ECS CLI Logging

## Overview

The following Proposal lays out a design and implementation plan for creating a user experience in the ECS CLI for getting container logs from CloudWatch.

### Use Cases

1. User has a known error and wants to find more info on it.
2. User is doing a deployment and wants to tail the logs and grep for errors.
3. User wants to quickly set up their CloudWatch Logs based upon the configuration specified in their docker compose file, and have the CLI creates any necessary log groups for them.
4. User wants to monitor their task/service, so they continually stream the logs.

## Phase 1 Solution
Top level `ecs-cli logs` command that will not use the docker compose file. This allows it to be used by a wide array of ECS customers, not just compose users. The command will allow customers to find logs for a given task.

```
ecs-cli logs --help
--follow                         Stream logs (continuously poll for updates)
--task-id                        [Required] View logs for a given Task ID
--task-def                      Required with Task ID if the task has been stopped already. Format: family:revision
--filter-pattern                        Substring to search for within the logs.
--container-name, -c     Filter logs for a given container definition
--since                           Filter logs in the last X minutes (can not be used with start time and end time)
--start-time                    Filter logs within a time frame, use with --end-time
--end-time                     Filter logs within a time frame, use with --end-time
--timestamps, -t           View time-stamps with the logs
```

```
ecs-cli logs --task-id d86079d1-6858-45e9-8ce2-1ba881c55c12 --time-stamps
Time-stamp            Message
2017-09-28 22:32:11 WordPress not found in /var/www/html - copying now...
2017-09-28 22:32:11 Complete! WordPress has been successfully copied to /var/www/html
2017-09-28 22:32:12 AH00558: apache2: Could not reliably determine the server's fully qualified domain name, using 172.17.0.3. Set the 'ServerName' directive globally to suppress this message
2017-09-28 22:32:12 AH00558: apache2: Could not reliably determine the server's fully qualified domain name, using 172.17.0.3. Set the 'ServerName' directive globally to suppress this message
2017-09-28 22:32:12 [Wed Sep 27 22:32:12.300422 2017] [mpm_prefork:notice] [pid 1] AH00163: Apache/2.4.10 (Debian) PHP/5.6.31 configured -- resuming normal operations
2017-09-28 22:32:12 [Wed Sep 27 22:32:12.300456 2017] [core:notice] [pid 1] AH00094: Command line: 'apache2 -D FOREGROUND'
```

### Implementation

- The logs implementation will not include any pagination- the command will return all logs corresponding to the specified search. We expect most users will be piping the output of the command to save it to a file, so this should not be a problem.
- If the user has not specified a log stream prefix in their task definition, then the command will fail and print an error message. Because without the log stream prefix set, we have no way of getting the logs for an individual task.
- For performance reasons, the command will only pull from *a single log group*. If the customer has not configured all of their container definitions to use the same log group, then the command will fail with an error and tell the customer they must re-run the command with the `--container-name` argument. This way, only 1 log group needs to be queried.

Work Flow:
1.	User gives Task ID
2.	Call Describe Tasks to get the TaskDef ARN (Skip this step if user provides Task Def)
3.	Call Describe Task Definition to get the Container Definitions.
4.	From Container Definitions, get the log configuration.
5.	Create a list of log streams that correspond to the correct task for the log group.
6.	Call FilterLogEvents on the log group to get the log events.
7.	Print log events.


## Compose Logs (Phase 2)
Phase 2 will be implemented in the future when we have time, it is lower priority than Phase 1, and thus Phase 2 may not be implemented for some time. We welcome the contributions of any customer who wishes to help start implementation of Phase 2 sooner.

### Configure Logs
- Log configuration using the docker-compose file is already supported
- Problem: Customer not required specify log stream prefix, however, we basically need log stream prefix to be specified because of how the ECS Agent sets the log stream name. If prefix is specified then it adds the container name and task ID to the log stream name (so we can use it to get the logs for each task). However, if a prefix is not specified, then the log stream name will for all intents and purposes be a random useless string (its an ID picked by the docker daemon on the instance, which from our point of view is meaningless).
    - The log stream will be named like this (by the ECS Agent): `prefix-name/container-name/ecs-task-id`

*Solution:* Existing ability to configure logs remains undisturbed, but add additional flag `--create-log-groups` that creates the necessary log group(s) in CloudWatch.
- The log configuration from the docker compose file will be read
- If user has not specified a log stream prefix, warn them that we are auto-setting it to a default value in their task definition.
    - *Additionally*, even if `--create-log-groups` is not specified, but we detect that the there is no prefix configured in their docker compose file (but log group and awslogs driver is specified), prefix will still be auto-set, and the user will be warned about this. This technically will break backwards compatibility- however, this risk is acceptable. It is very unlikely that ECS CLI users would actually desire to have their log streams named without a prefix. If no prefix is given, the ECS Agent sets the log stream name to be the container ID which was randomly generated by docker. Understanding this random ID requires logging into the underlying instances and retrieving info from the Docker Daemon. For all intents and purposes, the container ID is meaningless from a customer standpoint.
    - The ECS CLI is designed to simplify workflows and make it easier to understand ECS. Therefore, we should be opinionated and protect users from accidentally configuring there logs in a poor way. We can help protect users from the less useful, complicated, legacy behavior of ECS.
    - Additionally, the user will be warned that the default retention policy is to keep all log events forever, causing them to be charged for all time. They can change the policy in the CloudWatch Console or AWS CLI.

```
ecs-cli compose up --help
--create-log-groups                     Creates any necessary log groups in CloudWatch.
```

```
ecs-cli compose up --create-log-groups
INFO[0000] Creating Resources in CloudWatch for your logs.
WARN[0001] You have not specified a log stream prefix, auto-setting it to 'ecs-compose-'
WARN[0002] By default, CloudWatch will store your logs forever, it is recommended that you set a retention policy.
```

*Suggested Configuration:*
- If the user has not specified a log configuration in their compose file, then using the `--create-log-groups` command will fail and will print a help message with the suggested configuration. Here is one possible idea:
For Services:
```
awslogs-group: ${cluster name}/${service name}
```
For Tasks:
```
awslogs-group: ${cluster name}/${task def family}
```

### View Logs
- New Commands: `ecs-cli compose logs`, and `ecs-cli compose service logs`
- *Log command reads the configuration in user's docker compose file*

*Solution:* In docker-compose, and ECS task def, logs are configured per container definition. In docker-compose, these are called services and they must have names. Therefore, a user can view the logs per container definition. Since the agent will add the task ID to the log stream name, we can also list the logs for each task.

```
ecs-cli compose logs --help
--follow                         Stream logs (continuously poll for updates)
--task-id                        View logs for a given Task ID
--container-name, -c     View logs for a given container definition
--since                           View logs in the last X minutes (can not be used with start time and end time)
--start-time                    View logs within a time frame, use with --end-time
--end-time                     View logs within a time frame, use with --end-time
--time-stamps, -t           View time-stamps with the logs
--output, -o                    Output to a file
```

User's docker-compose file:
```
version: '2'
services:
  mysql:
    image: mysql
    cpu_shares: 100
    mem_limit: 524288000
    cap_add:
      - ALL
    logging:
      driver: awslogs
      options:
        awslogs-group: ecs-log-streaming
        awslogs-region: us-west-2
        awslogs-stream-prefix: mysql-logs
  wordpress:
    image: wordpress
    cpu_shares: 132
    mem_limit: 524288001
    ports:
      - "80:80"
    links:
      - mysql
    logging:
      driver: awslogs
      options:
        awslogs-group: ecs-log-streaming
        awslogs-region: us-west-2
        awslogs-stream-prefix: wordpress-logs
```

##### Examples

*User views logs for all MySQL Containers:*
- Outputs the logs for all containers running the given container definition. Ie if the user has 10 tasks running using this compose file, then the logs for all 10 of the mysql containers will be outputted. The output can be organized by the task ID.

*Implementation Details:*
- From the compose file, we know the log group for this container definition. We can call DescribeLogStreams to get a list of the log streams. FilterLogEvents can then be called with the list of LogStreams to get the log events. Each returned log event will have the log stream name associated with it- this will contain the task ID.

```
ecs-cli compose logs --container-name mysql --time-stamps
INFO[0000] Showing logs for all mysql containers
_______________________________________
Task: d86079d1-6858-45e9-8ce2-1ba881c55c12
_______________________________________
Time-stamp            Message
2017-09-28 22:32:11 WordPress not found in /var/www/html - copying now...
2017-09-28 22:32:11 Complete! WordPress has been successfully copied to /var/www/html
2017-09-28 22:32:12 AH00558: apache2: Could not reliably determine the server's fully qualified domain name, using 172.17.0.3. Set the 'ServerName' directive globally to suppress this message
2017-09-28 22:32:12 AH00558: apache2: Could not reliably determine the server's fully qualified domain name, using 172.17.0.3. Set the 'ServerName' directive globally to suppress this message
_______________________________________
Task: d86079d1-6858-45e9-8ce2-1ba881c55c12
______________________________________
Time-stamp            Message
2017-09-28 22:32:12 [Wed Sep 27 22:32:12.300422 2017] [mpm_prefork:notice] [pid 1] AH00163: Apache/2.4.10 (Debian) PHP/5.6.31 configured -- resuming normal operations
2017-09-28 22:32:12 [Wed Sep 27 22:32:12.300456 2017] [core:notice] [pid 1] AH00094: Command line: 'apache2 -D FOREGROUND'
```


*User views logs for a given task:*
- Outputs the logs for a given task ID
- The logs can be organized by the container name

*Implementation Details:*
- From the compose file, we know the log group for each container definition. We can call DescribeLogStreams to get a list of the log streams for each container definition. Each log stream will contain the Task ID in its name- so we can then call FilterLogEvents and use only the log streams for the given task ID as arguments. Each returned log event will have the log stream name associated with it- this will contain the container name.

```
ecs-cli compose logs --task-id --t
Container: MySql
_______________________
Time-stamp            Message
2017-09-28 22:32:11 WordPress not found in /var/www/html - copying now...
2017-09-28 22:32:11 Complete! WordPress has been successfully copied to /var/www/html
2017-09-28 22:32:12 AH00558: apache2: Could not reliably determine the server's fully qualified domain name, using 172.17.0.3. Set the 'ServerName' directive globally to suppress this message
2017-09-28 22:32:12 AH00558: apache2: Could not reliably determine the server's fully qualified domain name, using 172.17.0.3. Set the 'ServerName' directive globally to suppress this message
_______________________
Container: Wordpress
_______________________
Time-stamp            Message
2017-09-28 22:32:12 [Wed Sep 27 22:32:12.300422 2017] [mpm_prefork:notice] [pid 1] AH00163: Apache/2.4.10 (Debian) PHP/5.6.31 configured -- resuming normal operations
2017-09-28 22:32:12 [Wed Sep 27 22:32:12.300456 2017] [core:notice] [pid 1] AH00094: Command line: 'apache2 -D FOREGROUND'
```
