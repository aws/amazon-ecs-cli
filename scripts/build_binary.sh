#!/bin/bash
# Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You
# may not use this file except in compliance with the License. A copy of
# the License is located at
#
# 	http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is
# distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF
# ANY KIND, either express or implied. See the License for the specific
# language governing permissions and limitations under the License.

# Normalize to working directory being build root (up one level from ./scripts)
ROOT=$( cd "$( dirname "${BASH_SOURCE[0]}" )/.." && pwd )
cd "${ROOT}"

# Builds the ecs-cli binary from source in the specified destination paths.
mkdir -p $1

GIT_DIRTY=`git diff --quiet || echo '*'`
GIT_SHORT_HASH="$GIT_DIRTY"`git rev-parse --short=7 HEAD`
GOOS=$TARGET_GOOS CGO_ENABLED=0 go build -installsuffix cgo -a -ldflags "-s -X github.com/aws/amazon-ecs-cli/ecs-cli/modules/version.Version=development -X github.com/aws/amazon-ecs-cli/ecs-cli/modules/version.gitShortHash=$GIT_SHORT_HASH" -o $1/ecs-cli ./ecs-cli/

