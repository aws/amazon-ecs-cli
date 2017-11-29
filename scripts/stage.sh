#!/bin/bash
# Copyright 2014-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

set -e

DRYRUN=true
ALLOW_DIRTY=false

AWS_PROFILE=""
S3_BUCKET=""
S3_ACL_OVERRIDE=""

source $(dirname "${0}")/publish-functions.sh

usage() {
	echo "Usage: ${0} -b BUCKET [OPTIONS]"
	echo
	echo "This script is responsible for staging new versions of the Amazon ECS CLI."
	echo "Push the image (and its md5sum) to S3 with -latest, -VERSION, and -SHA"
	echo
	echo "Options"
	echo "  -d  true|false  Dryrun (default is true)"
	echo "  -p  PROFILE     AWS CLI Profile (default is none)"
	echo "  -b  BUCKET      AWS S3 Bucket"
	echo "  -a  ACL         AWS S3 Object Canned ACL (default is public-read)"
	echo "  -i  true|false  Allow dirty builds"
	echo "  -h              Display this help message"
}

while getopts ":d:p:b:a:i:h" opt; do
	case ${opt} in
		d)
			if [[ "${OPTARG}" == "false" ]]; then
				DRYRUN=false
			fi
			;;
		p)
			AWS_PROFILE="${OPTARG}"
			;;
		b)
			S3_BUCKET="${OPTARG}"
			;;
		a)
			S3_ACL_OVERRIDE="${OPTARG}"
			;;
		i)
			if [[ "${OPTARG}" == "true" ]]; then
				ALLOW_DIRTY=true
			fi
			;;
		\?)
			echo "Invalid option -${OPTARG}" >&2
			usage
			exit 1
			;;
		:)
			echo "Option -${OPTARG} requires an argument." >&2
			usage
			exit 1
			;;
		h)
			usage
			exit 0
			;;
	esac
done

if [ -z "${S3_BUCKET}" ]; then
	usage
	exit 1
fi

DIRTY_WARNING=$(cat <<EOW
***WARNING***
You currently have uncommitted or unstaged changes in your git repository.
The release build will not include those and the result may behave differently
than expected due to that. Please commit, stash, or remove all uncommitted or
unstaged files before creating a release build.
EOW
)
[ ! -z "$(git status --porcelain)" ] && echo "$DIRTY_WARNING"

CWD=$(pwd)
clean_directory=""
if ! ${ALLOW_DIRTY}; then
	clean_directory=$(mktemp -d)
	echo "Cloning to a clean directory ${clean_directory}"
	git clone --quiet "${CWD}" "${clean_directory}"
	cd "${clean_directory}"
	export ECS_RELEASE="cleanbuild"
fi

make docker-build

publish_binary() {
	platform=$1
	extension=$2
	artifact="bin/${platform}/ecs-cli${extension}"
	artifact_md5="$(mktemp)"
	md5sum "${artifact}" | sed 's/ .*//' > "${artifact_md5}"

	for tag in ${ARTIFACT_TAG_VERSION} ${ARTIFACT_TAG_SHA} ${ARTIFACT_TAG_LATEST}; do
		key_name="ecs-cli-${platform}-${tag}${extension}"
		echo "Publishing as ${key_name} with md5sum $(cat ${artifact_md5})"
		dexec s3_cp "${artifact}" "s3://${S3_BUCKET}/${key_name}"
		dexec s3_cp "${artifact_md5}" "s3://${S3_BUCKET}/ecs-cli-${platform}-${tag}.md5"
	done

	rm "${artifact_md5}"
}

publish_binary "linux-amd64" ""
publish_binary "darwin-amd64" ""
publish_binary "windows-amd64" ".exe"

cd "${CWD}"

if [[ -n "${clean_directory}" ]]; then
	echo "Removing ${clean_directory}"
	rm -rf "${clean_directory}"
fi
