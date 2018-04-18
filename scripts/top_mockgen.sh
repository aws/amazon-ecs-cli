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
declare -a MODULES
declare -a DONE
path="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
cd $path/..
for file in $(git ls-files ecs-cli/modules) ; do
    if grep -ql "//go:generate" $file ; then
        module=$(go list -f '{{ .ImportPath }}' ./$(dirname $file)) 
        MODULES+=("$module")
    fi
done

# :patrik, stackoverflow
function containsElement {
  local e match="$1"
  shift
  for e; do [[ "$e" == "$match" ]] && return 0; done
  return 1
}

function isAmongModules {
    local module
    for module in "${MODULES[@]}" ; do
        if [[ $1 = $module* ]] ; then
            echo $module
        fi
    done
}

function tryGen {
    local module=$1
    if containsElement "$module" ${DONE[@]} ; then
        return
    fi
    for import in $(go list -f '{{ join .TestImports " " }}' $module) ; do
        IFS='/' read -r -a fn <<<${import#*$module}
        local matched=$(isAmongModules $import)
        if [ ! -z $matched ] && [[ "$matched" != $module* ]] && containsElement "mock" ${fn[@]} ; then
            tryGen $matched
        fi
    done
    PATH=$path:$PATH go generate $module
    DONE+=("$module")
}

for module in "${MODULES[@]}" ; do
    tryGen $module
done
