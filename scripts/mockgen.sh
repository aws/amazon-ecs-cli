#!/bin/bash
# Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the
# "License"). You may not use this file except in compliance
#  with the License. A copy of the License is located at
#
#     http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is
# distributed on an "AS IS" BASIS, WITHOUT WARRANTIES OR
# CONDITIONS OF ANY KIND, either express or implied. See the
# License for the specific language governing permissions and
# limitations under the License.
#
# This script wraps the mockgen tool and inserts licensing information.

set -e
package=${1?Must provide package}
interfaces=${2?Must provide interface names}
outputfile=${3?Must provide an output file}

export PATH="${GOPATH//://bin:}/bin:$PATH"

wd=$(dirname "$0")
data=$(
cat << EOF
// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//     http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

$($wd/mockgen "$package" "$interfaces")
EOF
)

mkdir -p ./$(dirname $outputfile)
echo "$data" | goimports > "${outputfile}"
