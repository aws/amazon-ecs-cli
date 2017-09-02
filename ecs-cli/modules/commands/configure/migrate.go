// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package configure

import (
	"io/ioutil"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
)

func migrateWarning(config *config.CLIConfig) {
	dat, err := ioutil.ReadFile(path)
	if err != nil {
	}

}

var messageTemplate = `
Old config
-----------------------------------------------------
~/.ecs/config
{{.OldConfig}}

New configs
-----------------------------------------------------
~/.ecs/config
default: ecs
cluster-configurations:
  ecs:
  - cluster: {{.Cluster}}
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
`
