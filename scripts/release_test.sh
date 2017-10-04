#
# This scripts uploads run_commands.sh to the given ec2 instance and runs it.
# It then also runs the run_commands.sh on your local OS X Machine
#


# EXAMPLE:
# ./scripts/release_test.sh -c ReleaseTestCluster -r us-east-2
# -a $AWS_ACCESS_KEY_ID -s $AWS_SECRET_ACCESS_KEY -k MyFirstKeyPair
# -p ~/Downloads/KeyPair.pem -i ec2-18-220-215-243.us-east-2.compute.amazonaws.com
# -m $HOME/releasedir -h 4ff2bfcd -v 0.6.5

#default
DEFAULT_TEST_RESULT_DIR="~/ecs-cli-test-results"

usage() {
	echo "Usage: ${0}"
	echo
	echo "This script runs a series of ECS CLI commands during a release."
	echo "It also tests the cluster up command in all regions- verifying that the correct AMIs have been specified."
	echo "It is designed to provide sanity checking for newly built OSX and Linux Binaries."
	echo "However, it is called a sanity check for a reason, this script simply"
	echo "checks if the commands run exited with code 0."
	echo "Detailed test output can be examined manually in in the tesing dir on each machine."
	echo "(unless you have overridden the test output directory)."
	echo
	echo "Required Arguments:"
	echo "  -a  ACCESS KEY         The access key id for an AWS account."
	echo "  -s  SECRET KEY         The secret key for an AWS account."
	echo "  -k  KEY PAIR NAME      A keypair for the chosen region."
	echo "  -p  KEY PAIR PATH      The path to the key pair needed to ssh into the given ec2 instance."
	echo "  -i  INSTANCE URL       The public DNS for the ec2 instance created for linux testing."
	echo "  -m  OS X TEST DIR      The directory on the OS X Machine to store test info and perform the tests."
	echo "  -l  Linux TEST DIR     [Optional] The directory on the Linux Machine to store test info. Defaults to $DEFAULT_TEST_RESULT_DIR."
	echo "  -h  hash               The 7 char git hash for this CLI version."
	echo "  -v  version            The version number for the CLI. Ex: 0.6.5"
	echo "  -h                     Display this help message"
}

# ARGS: cluster c, region r, access a, secret s, keypair k , keypath p, instance_url i, mac_dir,
while getopts ":c:r:a:s:k:p:i:m:l:v:h:" opt; do
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
		m)
			mac_dir="${OPTARG}"
			;;
		l)
			linux_dir="${OPTARG}"
			;;
		v)
			version="${OPTARG}"
			;;
		h)
			hash="${OPTARG}"
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

if [ -z "${linux_dir}" ]; then
	linux_dir=$DEFAULT_TEST_RESULT_DIR
fi

# run locally on mac first
cp $(dirname "${0}")/run_commands.sh $mac_dir/
cp $(dirname "${0}")/../integration-tests/docker-compose.yml $mac_dir/
cd $mac_dir
curl -o ecs-cli https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-latest
curl -o ecs-cli-"${version}" https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-v"${version}"
curl -o ecs-cli-"${hash}" https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-darwin-amd64-"${hash}"
chmod +x run_commands.sh
cluster=${cluster} region=${region} access=${access} secret=${secret} keypair=${keypair} keypath=${keypath} instance_url=${instance_url} TEST_RESULT_DIR=${mac_dir} version=${version} hash=${hash} release='yes' ./run_commands.sh;

cd -



# Run on Linux
scp -i $keypath $(dirname "${0}")/../integration-tests/docker-compose.yml "ec2-user@${instance_url}":~/
scp -i $keypath $(dirname "${0}")/run_commands.sh "ec2-user@${instance_url}":~/
# ARGS: cluster c, region r, access a, secret s, keypair k , keypath p, instance_url i, gitname u, branch b
ssh -i $keypath "ec2-user@${instance_url}" "chmod +x run_commands.sh"
ssh -i $keypath "ec2-user@${instance_url}" "cluster=${cluster} region=${region} access=${access} secret=${secret} keypair=${keypair} keypath=${keypath} instance_url=${instance_url} TEST_RESULT_DIR=${linux_dir} version=${version} hash=${hash} release='yes' linux='yes' ./run_commands.sh"
