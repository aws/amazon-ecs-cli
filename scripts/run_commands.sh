#
# This scripts runs on the ec2 instance. It is uploaded to the EC2 instance by
# test_commands.sh
# The necessary env variables will be set by test_commands.sh using the
# command line arguments provided to it.
#

declare -i FAILURES
FAILURES=0

RED='\033[1;31m'
GREEN='\033[1;32m'
LRED='\033[0;31m'
LGREEN='\033[0;32m'
NC='\033[0m' # No Color

expect_success() {
	if "${@}" &>> ~/ecs-cli-test-results/test_output.txt; then
		echo -e "${GREEN}SUCCEEDED${NC}: ${LGREEN}${@}${NC}"
	else
		echo "-------"
		echo -e "${RED}FAILED${NC}: ${LRED}${@}${NC}"
		echo "-------"
		FAILURES+=1
	fi
}

# used for commands that whose arguments might contain sensitive information
expect_success_no_log() {
	if "${@}" &>> ~/ecs-cli-test-results/test_output.txt; then
		echo -e "${GREEN}SUCCEEDED${NC}: ${LGREEN}${1} ${2}${NC}"
	else
		echo "-------"
		echo -e "${RED}FAILED${NC}: ${LRED}${1} ${2}${NC}"
		echo "-------"
		FAILURES+=1
	fi
}

# install git and go
echo "TYPE y|yes then enter to proceed."
sudo yum install git go >> ~/ecs-cli-test-results/test_log.txt
# have to respond yes to prompt
# get CLI
export GOPATH="$HOME/go"
go get github.com/aws/amazon-ecs-cli >> ~/ecs-cli-test-results/test_log.txt
cd $GOPATH/src/github.com/aws/amazon-ecs-cli
url="https://github.com/"
url+=$gitname
url+="/amazon-ecs-cli.git"
git remote add fork $url &>> ~/ecs-cli-test-results/test_log.txt
git fetch fork &>> ~/ecs-cli-test-results/test_log.txt
git checkout "fork/${branch}"
make build

rm ~/ecs-cli-test-results/test_output.txt # clean up from past tests

# test commands

# configure
expect_success_no_log ./bin/local/ecs-cli configure --region $region --access-key $access --secret-key $secret --cluster $cluster
# up
expect_success ./bin/local/ecs-cli up --capability-iam --keypair $keypair --size 1 --instance-type t2.medium --force
# create a service
expect_success ./bin/local/ecs-cli compose --file ./integration-tests/docker-compose.yml service up
# take down service
expect_success ./bin/local/ecs-cli compose --file ./integration-tests/docker-compose.yml service down
# create a task
expect_success ./bin/local/ecs-cli compose --file ./integration-tests/docker-compose.yml up
# take down task
expect_success ./bin/local/ecs-cli compose --file ./integration-tests/docker-compose.yml down
# take down cluster
expect_success ./bin/local/ecs-cli down --force


# Sanity Check- search output file for error messages
cat ~/ecs-cli-test-results/test_output.txt | grep -i "err" > ~/ecs-cli-test-results/errors.txt
cat ~/ecs-cli-test-results/test_output.txt | grep -i "fatal" >> ~/ecs-cli-test-results/errors.txt
echo -e "${RED}"
cat ~/ecs-cli-test-results/errors.txt
echo -e "${NC}"

echo "ALL TESTS COMPLETE"
if [[ $FAILURES -eq 0 ]]; then
	echo "--------------------"
	echo -e "${GREEN}ALL TESTS SUCCEEDED.${NC}"
	echo "--------------------"
else
	echo "--------------------"
	echo -e "${RED}${FAILURES} TEST FAILURES${NC}"
	echo "--------------------"
fi
