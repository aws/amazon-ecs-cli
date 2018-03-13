// NOTE: DEPRECATED. These structs are only left here so that
// we can read old ini based config files for customers who have
// still been using older versions of the CLI. All new config files
// will be written in the YAML format.

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

package config

const (
	ecsSectionKey = "ecs"
)

// iniCLIConfig is the struct used to map to the ini config.
// This is to allow us to read old ini based config files
// CliConfig has been updated to use the yaml annotations
type iniCLIConfig struct {
	*IniSectionKeys `ini:"ecs"`
}

// SectionKeys is the struct embedded in iniCLIConfig. It groups all the keys in the 'ecs' section in the ini file.
type IniSectionKeys struct {
	Cluster                  string `ini:"cluster"`
	AwsProfile               string `ini:"aws_profile"`
	Region                   string `ini:"region"`
	AWSAccessKey             string `ini:"aws_access_key_id"`
	AWSSecretKey             string `ini:"aws_secret_access_key"`
	ComposeProjectNamePrefix string `ini:"compose-project-name-prefix"`
	ComposeServiceNamePrefix string `ini:"compose-service-name-prefix"`
	CFNStackNamePrefix       string `ini:"cfn-stack-name-prefix"`
}
