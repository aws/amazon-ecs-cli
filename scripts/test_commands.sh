#
# This scripts uploads run_commands.sh to the given ec2 instance and runs it.
#


# EXAMPLE:
# ./scripts/test_commands.sh -c IntegrationTestCluster -r us-east-2
# -a $AWS_ACCESS_KEY_ID -s $AWS_SECRET_ACCESS_KEY -k MyFirstKeyPair
# -p ~/Downloads/KeyPair.pem -i ec2-18-220-215-243.us-east-2.compute.amazonaws.com

TEST_RESULT_DIR="~/ecs-cli-test-results"

usage() {
	echo "Usage: ${0}"
	echo
	echo "This script runs a series of ECS CLI commands on a fresh EC2 instance."
	echo "It is designed to provide integration tests for the ECS CLI."
	echo "However, it is called a sanity check for a reason, this script simply"
	echo "checks if the commands run exited with code 0."
	echo "If -u (GitHub username) and -b (Branch) options are specified,"
	echo "then the script pulls and tests a branch on Github."
	echo "Otherwise it tests local changes."
	echo "More detailed test output can be examined manually in"
	echo "$TEST_RESULT_DIR/test_output.txt, on the ec2 instance."
	echo
	echo "Required Arguments:"
	echo "  -c  CLUSTER            A name for the cluster that will be created/used in the tests."
	echo "  -r  REGION             Region to create AWS resources in."
	echo "  -a  ACCESS KEY         The access key id for an AWS account."
	echo "  -s  SECRET KEY         The secret key for an AWS account."
	echo "  -k  KEY PAIR NAME      A keypair for the chosen region."
	echo "  -p  KEY PAIR PATH      The path to the key pair needed to ssh into the given ec2 instance."
	echo "  -i  INSTANCE URL       The public DNS for the ec2 instance created for testing."
	echo "  -u  GITHUB USERNAME    [OPTIONAL] The testers github username."
	echo "  -b  BRANCH             [OPTIONAL] The branch on github that will be tested."
	echo "  -h                     Display this help message"
}

# ARGS: cluster c, region r, access a, secret s, keypair k , keypath p, instance_url i, gitname u, branch b
while getopts ":c:r:a:s:k:p:i:u:b:" opt; do
	case ${opt} in
		c)
			cluster="${OPTARG}"
			;;
		r)
			region="${OPTARG}"
			;;
		a)
			access="${OPTARG}"
			;;
		s)
			secret="${OPTARG}"
			;;
		k)
			keypair="${OPTARG}"
			;;
		p)
			keypath="${OPTARG}"
			;;
		i)
			instance_url="${OPTARG}"
			;;
		u)
			gitname="${OPTARG}"
			;;
		b)
			branch="${OPTARG}"
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

if [ -z "${instance_url}" ]; then
	usage
	exit 1
fi

if [ -z "${gitname}" ]; then
	#  Not testing a branch, test local changes instead
	make docker-build
	scp -i $keypath $(dirname "${0}")/../bin/linux-amd64/ecs-cli "ec2-user@${instance_url}":~/
fi

scp -i $keypath $(dirname "${0}")/../integration-tests/docker-compose.yml "ec2-user@${instance_url}":~/
scp -i $keypath $(dirname "${0}")/run_commands.sh "ec2-user@${instance_url}":~/
# ARGS: cluster c, region r, access a, secret s, keypair k , keypath p, instance_url i, gitname u, branch b
ssh -i $keypath "ec2-user@${instance_url}" "chmod +x run_commands.sh"
ssh -i $keypath "ec2-user@${instance_url}" "cluster=${cluster} region=${region} access=${access} secret=${secret} keypair=${keypair} keypath=${keypath} instance_url=${instance_url} gitname=${gitname} branch=${branch} TEST_RESULT_DIR=${TEST_RESULT_DIR} ./run_commands.sh"
