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
	"fmt"
	"html/template"
	"io/ioutil"
	"os"
	"strings"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
)

func maskSecret(secret string) (out string) {
	if secret != "" {
		mask := strings.Repeat("*", len(secret)-2)
		out = strings.Replace(secret, secret[:len(secret)-2], mask, 1)
	}
	return out
}

func hideCreds(cliConfig *config.CLIConfig) {
	if cliConfig.AWSSecretKey != "" {
		cliConfig.AWSSecretKey = maskSecret(cliConfig.AWSSecretKey)
	}
}

func hideCredsOldFile(data string) string {
	safeData := ""
	lines := strings.Split(data, "\n")
	for _, line := range lines {
		if strings.Contains(line, "aws_secret_access_key") {
			parts := strings.Split(line, "=")
			if strings.TrimSpace(parts[1]) != "" {
				line = fmt.Sprintf("%v= %v", parts[0], maskSecret(parts[1]))
			}
		}
		safeData += line
		safeData += "\n"
	}

	return safeData
}

func migrateWarning(cliConfig config.CLIConfig) error {
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

	hideCreds(&cliConfig)
	oldConfig = hideCredsOldFile(oldConfig)

	optionalFields := ""

	if cliConfig.ComposeServiceNamePrefix != "" {
		optionalFields += "compose-service-name-prefix: " + cliConfig.ComposeServiceNamePrefix + "\n"
	}

	if cliConfig.CFNStackName != "" {
		optionalFields += "cfn-stack-name: " + cliConfig.CFNStackName + "\n"
	}

	// format template
	data := map[string]interface{}{
		"OldConfig":       oldConfig,
		"Cluster":         cliConfig.Cluster,
		"Region":          cliConfig.Region,
		"AWSAccessKey":    cliConfig.AWSAccessKey,
		"AWSSecretKey":    cliConfig.AWSSecretKey,
		"Optional_Fields": optionalFields,
	}

	t := template.Must(template.New("message").Parse(messageTemplate))
	return t.Execute(os.Stdout, data)
}

var messageTemplate = `
Old configuration file
-----------------------------------------------------
~/.ecs/config
{{.OldConfig}}

New configuration files
-----------------------------------------------------
~/.ecs/config
version: v1
default: default
cluster-configurations:
  default:
    cluster: {{.Cluster}}
    region: {{.Region}}
    {{.Optional_Fields}}

{{if .AWSSecretKey}}
~/.ecs/credentials
default: default
credentials:
  default:
    aws_access_key_id: {{.AWSAccessKey}}
    aws_secret_access_key: {{.AWSSecretKey}}
{{end}}

[WARN] Please read the following changes carefully: http://docs.aws.amazon.com/AmazonECS/latest/developerguide/ECS_CLI_Configuration.html
- The option --compose-project-name-prefix has been removed (name used for create task definition: <compose_project_name_prefix> + <project_name>). You can specify your desired name with the --project-name option.
- The --compose-service-name-prefix option has been deprecated (name used for create service: <compose_service_name_prefix> + <project_name>). This field can still be configured; however, if it is not configured there is no longer a default value assigned.
- The --cfn-stack-name-prefix option has been removed. To use an existing CloudFormation stack, please specify the full stack name using the --cfn-stack-name option; otherwise, the stack name defaults to amazon-ecs-cli-setup-<cluster_name>.
- Storing an AWS Profile name in the config is no longer supported, please use the --aws-profile flag inline instead.

Are you sure you want to migrate[y/n]?
`
