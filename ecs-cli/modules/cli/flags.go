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

package cli

// Flag names used by the cli.
// TODO: These need a better home.
const (
	AccessKeyFlag          = "access-key"
	SecretKeyFlag          = "secret-key"
	RegionFlag             = "region"
	AwsRegionEnvVar        = "AWS_REGION"
	AwsDefaultRegionEnvVar = "AWS_DEFAULT_REGION"
	ProfileFlag            = "profile"
	ClusterFlag            = "cluster"
	VerboseFlag            = "verbose"
	RegistryIdFlag         = "registry-id"
	FromFlag               = "from"
	ToFlag                 = "to"

	ComposeProjectNamePrefixFlag         = "compose-project-name-prefix"
	ComposeProjectNamePrefixDefaultValue = "ecscompose-"
	ComposeServiceNamePrefixFlag         = "compose-service-name-prefix"
	ComposeServiceNamePrefixDefaultValue = ComposeProjectNamePrefixDefaultValue + "service-"
	CFNStackNamePrefixFlag               = "cfn-stack-name-prefix"
	CFNStackNamePrefixDefaultValue       = "amazon-ecs-cli-setup-"
)
