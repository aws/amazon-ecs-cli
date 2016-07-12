#!/bin/bash
# Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
#
# Licensed under the Apache License, Version 2.0 (the "License"). You may
# not use this file except in compliance with the License. A copy of the
# License is located at
#
#	http://aws.amazon.com/apache2.0/
#
# or in the "license" file accompanying this file. This file is distributed
# on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
# express or implied. See the License for the specific language governing
# permissions and limitations under the License.
#
# This script generates a file in go with the license contents as a constant

set -e
outputfile=${1?Must provide an output file}
inputfile="$(<../../LICENSE)"

for user in ./../vendor/github.com/*; do
  for repo in $user/*; do
    inputfile+=$'\n'"***"$'\n'"$repo"$'\n\n'
    if [ -f $repo/LICENSE* ]; then
      inputfile+="$(<$repo/LICENSE*)"$'\n'
    elif [ -f $repo/COPYING* ]; then
      inputfile+="$(<$repo/COPYING*)"$'\n'
    fi;
  done;
done;

cat << EOF > "${outputfile}"
// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package license

const License = \`$inputfile\`
EOF
