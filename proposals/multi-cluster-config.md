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
ECS CLI currently only support one cluster configuration which results in a number of issues for the users:
1. When customer switch clusters or regions within the same account, they need to reconfigure the setup.
1. When customer is running two Jenkins builds simultaneously, the config file gets reconfigured while one job is still processing.
1. When customers work as a team, each person has to run `ecs-cli configure` to configure the correct cluster and region.
1. Everything (cluster name, region, creds) should be configurable via arguments or environment variables.

# Solution
When users run `ecs-cli configure`, ECS CLI will configure their cluster configuration in `~/.ecs/ecs.config`
```yaml
default: beta
cluster-configurations:
  beta:
  - cluster: cli-demo
  - region: us-west-2
  gamma:
  - cluster: cli-demo2
  - region: us-west-1
  prod:
  - ...
```

When users run `ecs-cli configure profile`, ECS CLI will configure their credentials in `~/.ecs/ecs.profile`
```yaml
default: my_profile
ecs_profiles:
  my_profile:
  - aws_access_key_id     : **********
  - aws_secret_access_key : **********
  dev_profile:
  - aws_access_key_id     : **********
  - aws_secret_access_key : **********
  team_profile:
  - ...
```

Other commands will have additional flags
* **--cluster-config**
* --cluster
* --region
* --aws-profile (this is --profile today)
* **--ecs-profile**

# Workflow
### Single cluster first run
1. Configure ECS credentials (Stores in `~/.ecs/ecs.profile`)
   ```
   ecs-cli configure  profile --access-key ********** --secret-key ********** --profile-name my_profile
   ```
1. Configure cluster (Stores in `~/.ecs/ecs.config`)
   ```
   ecs-cli configure --cluster cli-demo --region us-west-2 --config-name beta_config
   ```
1. Spin up a cluster
   ```
   ecs-cli up
   ```
1. Run tasks
   ```
   ecs-cli compose up
   ```

### Single cluster second run
1. Configure another ECS profile
   ```
   ecs-cli configure  profile --access-key ********** --secret-key ********** --profile-name gamma_profile
   ```
1. Set default ECS profile
   ```
   ecs-cli configure profile default --profile-name gamma_profile
   ```
1. Configure another cluster
   ```
   ecs-cli configure --cluster cli-demo --region us-west-2 --config-name gamma_config
   ```
1. Set default cluster
   ```
   ecs-cli configure default --config-name gamma_config
   ```
1. Spin up a cluster
   ```
   ecs-cli up
   ```
1. Run tasks
   ```
   ecs-cli compose up
   ```

### Multiple clusters
1. Configure ECS profile
   ```
   ecs-cli configure  profile --access-key ********** --secret-key ********** --profile-name my_profile
   ```
1. Configure and spin up the first cluster
   ```
   ecs-cli configure --cluster cli-demo --region us-west-2 --config-name beta_config

   ecs-cli up --cluster-config beta_config
   ```
1. Configure and spin up a second cluster
   ```
   ecs-cli configure --cluster cli-demo2 --region us-west-2 --config-name prod_config

   ecs-cli up --cluster-config prod_config
   ```
1. Run tasks on different clusters
   ```
   ecs-cli compose up --cluster-config beta_config

   ecs-cli compose up --cluster-config beta_config
   ```
1. Do the same on a different account
   ```
   ecs-cli configure  profile --access-key ********** --secret-key ********** --profile-name prod_profile

   ecs-cli up --cluster-config prod_config --ecs-profile prod_profile

   ecs-cli compose up --cluster-config prod_config --ecs-profile prod_profile
   ```
1. Do the same with AWS profile
   ```
   ecs-cli up --cluster-config prod_config --aws-profile default

   ecs-cli compose up --cluster-config prod_config --aws-profile default
   ```

### Ad-hoc
Does not need to configure cluster
1. Configure ECS profile
   ```
   ecs-cli configure  profile --access-key ********** --secret-key ********** --profile-name my_profile
   ```
1. Cluster PS
   ```
   ecs-cli ps --cluster cli-demo --region us-west-2 --ecs-profile my_profile
   ```
1. Do the same with AWS profile
   ```
   ecs-cli ps --cluster cli-demo --region us-west-2 --aws-profile default
   ```
