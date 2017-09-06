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
	"html/template"
	"io/ioutil"
	"os"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
)

func migrateWarning(cliConfig *config.CLIConfig) error {
	var oldConfig string
	dest, err := config.NewDefaultDestination()
	if err != nil {
		return err
	}
	dat, err := ioutil.ReadFile(config.ConfigFilePath(dest))
	if err != nil {
		return err
	}
	oldConfig = string(dat)

	// format template
	data := map[string]interface{}{
		"OldConfig":                oldConfig,
		"Cluster":                  cliConfig.Cluster,
		"Region":                   cliConfig.Region,
		"CFNStackName":             cliConfig.CFNStackName,
		"ComposeServiceNamePrefix": cliConfig.ComposeServiceNamePrefix,
		"AWSAccessKey":             cliConfig.AWSAccessKey,
		"AWSSecretKey":             cliConfig.AWSSecretKey,
	}

	t := template.Must(template.New("message").Parse(messageTemplate))
	if err := t.Execute(os.Stdout, data); err != nil {
		return err
	}

	return nil

}

var messageTemplate = `
Old config
-----------------------------------------------------
~/.ecs/config
{{.OldConfig}}

New configs
-----------------------------------------------------
~/.ecs/config
default: default
cluster-configurations:
  default:
  - cluster: {{.Cluster}}
  - region: {{.Region}}
  - cfn-stack-name: {{.CFNStackName}}
  - compose-service-name-prefix: {{.ComposeServiceNamePrefix}}

~/.ecs/credentials
default: default
credentials:
  default:
  - aws_access_key_id: {{.AWSAccessKey}}
  - aws_secret_access_key: {{.AWSSecretKey}}

[WARN] Please read the following changes carefully: <link to documentation>
- compose-project-name-prefix and compose-service-name-prefix are deprecated, you can continue to specify your desired names using the runtime flag --project-name
- cfn-stack-name-prefix no longer exists, if you wish to continue to use your existing CFN stack, please specify the full stack name, otherwise it will be defaulted to  amazon-ecs-cli-setup-<cluster_name>

Are you sure you wish to migrate[y/n]?
`
