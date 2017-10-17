<!--
Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

Licensed under the Apache License, Version 2.0 (the "License"). You may
not use this file except in compliance with the License. A copy of the
License is located at

http://aws.amazon.com/apache2.0/

or in the "license" file accompanying this file. This file is distributed
on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
express or implied. See the License for the specific language governing
permissions and limitations under the License.
-->

# Introduction
As the CLI moves to multiple cluster configurations, the prefixes in the old configuration will be removed because:
1. Bad user experience and not really a cluster configuration.
1. Every new customer needs to understand what the prefixes are for in order to know the names of the resources created by ECS CLI
1. Customers do not have the flexibility to use the resource name they wish to have unless they set the prefixes to empty string.

The 3 prefixes that are used in the configuration are:
* **compose-project-name-prefix** - name used for create task definition (<compose_project_name_prefix> + <project_name>)
* **compose-service-name-prefix** - name used for create service (<compose_service_name_prefix> + <project_name>)
* **cfn-stack-name-prefix** - name used for creating CFN stacks (<cfn_stack_name_prefix> + <cluster_name>)

# Solution
When users run `ecs-cli configure`, the prefixes will be removed from the config.
```yaml
default: my_config
cluster-configurations:
  my_config:
  - cluster: demo
  - region: us-west-2
```

When CLI detects that there is a different `cfn-stack-name` configured, the `cfn-stack-name` will be saved in the config, otherwise will default to `amazon-ecs-cli-setup-<cluster_name>`, similar to `compose-service-name-prefix` will be saved in the config if present.
```yaml
default: my_config
cluster-configurations:
  my_config:
  - cluster: demo
  - region: us-west-2
  - cfn-stack-name: amazon-ecs-cli-demo
  - compose-service-name-prefix: ecscompose-service-
```

# Workflow
Assume project name is "myApp"
### Existing customer first run
1. Config help
   ```
   $ ecs-cli configure --help
   
   COMMANDS:
   migrate    Moves your old configuration to new configuration
   
   OPTIONS:
   --cluster value, -c value            Specifies the ECS cluster name to use. If the cluster does not exist, it is created when you try to add resources to it with the ecs-cli up command. [$ECS_CLUSTER]
   --region value, -r value             Specifies the AWS region to use. If the AWS_REGION environment variable is set when ecs-cli configure is run, then the AWS region is set to the value of that environment variable. [$AWS_REGION]
   --cfn-stack-name value               [Optional] Specifies the name of AWS CloudFormation stack created on ecs-cli up. (default: "amazon-ecs-cli-setup-<cluster-name>")
   --compose-service-name-prefix value  [Deprecated] Specifies the prefix added to an ECS service created from a compose file. Format <prefix><project-name>. (default to empty)
   
   $ ecs-cli configure migrate --help
   --force                      [Optional] Force move your old configuration to new configuration

   ```

1. Migrate config with prompts
   ```
   $ ecs-cli configure migrate
   Old config
   -----------------------------------------------------
   ~/.ecs/config
   [ecs]
   cluster = cli-demo
   aws_profile =
   region = us-west-2
   aws_access_key_id  = *********
   aws_secret_access_key  = *********
   compose-project-name-prefix = ecscompose-
   compose-service-name-prefix = ecscompose-service-
   cfn-stack-name-prefix       = amazon-ecs-cli-setup-

   New configs
   -----------------------------------------------------
   ~/.ecs/config
   default: ecs
   cluster-configurations:
     ecs:
     - cluster: cli-demo
     - region: us-west-2
     - cfn-stack-name: amazon-ecs-cli-setup-cli-demo
     - compose-service-name-prefix: ecscompose-service-

   ~/.ecs/credentials
   default: ecs
   credentials:
     ecs:
     - aws_access_key_id: *********
     - aws_secret_access_key: *********

   [WARN] Please read the following changes carefully: <link to documentation>
   - compose-project-name-prefix and compose-service-name-prefix are deprecated, you can continue to specify your desired names using the runtime flag --project-name
   - cfn-stack-name-prefix no longer exists, if you wish to continue to use your existing CFN stack, please specify the full stack name, otherwise it will be defaulted to  amazon-ecs-cli-setup-<cluster_name>

   Are you sure you wish to migrate[y/n]?
   y
   ```
1. Run tasks
   ```
   $ ecs-cli compose up
   [INFO] RegisterTaskDefinition, TaskDefinitionName = myApp
   [INFO] StartTask

   $ ecs-cli compose ps
   [INFO] TaskDefinitionName = myApp

   $ ecs-cli compose down
   [INFO] TaskDefinitionName = myApp
   ```
1. Update service
   ```
   $ ecs-cli compose service up
   [WARN] `compose-service-name-prefix` is being deprecated in version 1.1, please run `ecs-cli configure` without the prefix and `ecs-cli compose --project-name ecscompose-service-myApp service up` to manage existing services
   [INFO] RegisterTaskDefinition, TaskDefinitionName = myApp

   $ ecs-cli compose service ps
   [WARN] `compose-service-name-prefix` is being deprecated in version 1.1, please run `ecs-cli configure` without the prefix and `ecs-cli compose --project-name ecscompose-service-myApp service ps` to manage existing services
   [INFO] ServiceName = ecscompose-service-myApp

   $ ecs-cli compose service down
   [WARN] `compose-service-name-prefix` is being deprecated in version 1.1, please run `ecs-cli configure` without the prefix and `ecs-cli compose --project-name ecscompose-service-myApp service down` to manage existing services
   [INFO] ServiceName = ecscompose-service-myApp
   ```
### Existing customer second run
1. Force migrating config, already knows what is going to happen
   ```
   $ ecs-cli configure migrate --force
   [WARN] Please read the following changes carefully: <link to documentation>
   - compose-project-name-prefix and compose-service-name-prefix is deprecated, you can continue to specify your desired names using the runtime flag --project-name
   - cfn-stack-name-prefix no longer exists, if you wish to continue to use your existing CFN stack, please specify the full stack name, otherwise it will be defaulted to    amazon-ecs-cli-setup-<cluster_name>
   ```
1. Re-configure without prefixes
   ```
   $ ecs-cli configure --cluster cli-demo --region us-west-2
   Saved config in ~/.ecs/config
   default: ecs
   cluster-configurations:
     ecs:
     - cluster: cli-demo
     - region: us-west-2
     - cfn-stack-name: amazon-ecs-cli-setup-cli-demo
   ```
1. Update existing cluster_name
   ```
   $ ecs-cli scale 5
   [INFO] CreateCluster, Cluster = cli-demo
   [INFO] CFNStackName = amazon-ecs-cli-setup-cli-demo

   $ ecs-cli ps
   ```
1. Create new task
   ```
   $ ecs-cli compose ps
   [INFO] TaskDefinitionName = myApp
   ```
1. Stop existing tasks running
   ```
   $ ecs-cli compose --project-name ecscompose-myApp ps
   [INFO] TaskDefinitionName = ecscompose-myApp

   $ ecs-cli compose --project-name ecscompose-myApp down
   [INFO] TaskDefinitionName = ecscompose-myApp
1. Create new service
   ```
   $ ecs-cli compose service up
   [INFO] ServiceName = myApp
   ```
1. Delete existing service
   ```
   $ ecs-cli compose --project-name ecscompose-service-myApp ps
   [INFO] ServiceName = ecscompose-service-myApp

   $ ecs-cli compose --project-name ecscompose-service-myApp service down
   [INFO] ServiceName = ecscompose-service-myApp
   ```
