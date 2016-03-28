#!/bin/bash
# Copyright 2015 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

set -e

vendor_github() {
  pushd .
  mkdir -p ./ecs-cli/vendor/src/github.com/$1/$2
  cd ./ecs-cli/vendor/src/github.com/$1
  git clone https://github.com/$1/$2
  cd $2
  git checkout "$3"
  rm -rf ./.git
  popd
}

rm -rf ./ecs-cli/vendor

vendor_github aws aws-sdk-go 6876e9922ff299adf36e43e04c94820077968b3b
vendor_github jmespath go-jmespath 0b12d6b521d83fc7f755e7cfc1b1fbdd35a01a74
vendor_github vaughan0 go-ini a98ad7ee00ec53921f08832bc06ecf7fd600e6a1
vendor_github Sirupsen logrus 418b41d23a1bf978c06faea5313ba194650ac088
vendor_github go-ini ini e8c222fea70c6c03bde4f0577c93965e7f91d417
vendor_github codegangsta cli a65b733b303f0055f8d324d805f393cd3e7a7904
vendor_github kylelemons go-gypsy 42fc2c7ee9b8bd0ff636cd2d7a8c0a49491044c5
vendor_github golang mock 06883d979f10cc178f2716846215c8cf90f9e363
