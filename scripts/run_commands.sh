# This scripts runs on the ec2 instance and it used for integration testing.
# It is uploaded to the EC2 instance by test_commands.sh.
# The full output of the test script is stored in $TEST_RESULT_DIR.
# The necessary env variables will be set by test_commands.sh using the
# command line arguments provided to it.

declare -i FAILURES
FAILURES=0

RED='\033[1;31m'
GREEN='\033[1;32m'
LRED='\033[0;31m'
LGREEN='\033[0;32m'
NC='\033[0m' # No Color

expect_success() {
	if "${@}" &>> $TEST_RESULT_DIR/test_output.txt; then
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
	if "${@}" &>> $TEST_RESULT_DIR/test_output.txt; then
		echo -e "${GREEN}SUCCEEDED${NC}: ${LGREEN}${1} ${2}${NC}"
	else
		echo "-------"
		echo -e "${RED}FAILED${NC}: ${LRED}${1} ${2}${NC}"
		echo "-------"
		FAILURES+=1
	fi
}

rm $TEST_RESULT_DIR/test_output.txt # clean up from past tests (in case this instance has been used before)
# create the results directory and file; in case this instance is being used for the first time
mkdir $TEST_RESULT_DIR/
touch $TEST_RESULT_DIR/test_output.txt

if ! [ -z "${gitname}" ]; then
	# Not testing local changes, testing a branch instead
	echo "TYPE y|yes to proceed and install git and go on the EC2 instance."
	sudo yum install git go >> $TEST_RESULT_DIR/test_log.txt # yum requires the user to confirm the install
	# have to respond yes to prompt
	# get CLI
	export GOPATH="$HOME/go"
	go get github.com/aws/amazon-ecs-cli >> $TEST_RESULT_DIR/test_log.txt
	cd $GOPATH/src/github.com/aws/amazon-ecs-cli
	url="https://github.com/"
	url+=$gitname
	url+="/amazon-ecs-cli.git"
	git remote add fork $url &>> $TEST_RESULT_DIR/test_log.txt
	git fetch fork &>> $TEST_RESULT_DIR/test_log.txt
	git checkout "fork/${branch}"
	make build

	cd bin/local/
fi

# test commands

# configure
expect_success_no_log ./ecs-cli configure --region $region --access-key $access --secret-key $secret --cluster $cluster
# up
expect_success ./ecs-cli up --capability-iam --keypair $keypair --size 1 --instance-type t2.medium --force
# create a service
expect_success ./ecs-cli compose --file ./integration-tests/docker-compose.yml service up
# ps on th service
expect_success ./ecs-cli compose --file ./integration-tests/docker-compose.yml service ps
# take down service
expect_success ./ecs-cli compose --file ./integration-tests/docker-compose.yml service down
# create a task
expect_success ./ecs-cli compose --file ./integration-tests/docker-compose.yml up
# ps the task
expect_success ./ecs-cli compose --file ./integration-tests/docker-compose.yml ps
# take down task
expect_success ./ecs-cli compose --file ./integration-tests/docker-compose.yml down
# take down cluster
expect_success ./ecs-cli down --force


# Sanity Check- search output file for error messages
cat $TEST_RESULT_DIR/test_output.txt | grep -i "err" > $TEST_RESULT_DIR/errors.txt
cat $TEST_RESULT_DIR/test_output.txt | grep -i "fatal" >> $TEST_RESULT_DIR/errors.txt
echo -e "${RED}"
cat $TEST_RESULT_DIR/errors.txt
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
