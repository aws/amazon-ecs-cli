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
	sudo yum -y install git go >> $TEST_RESULT_DIR/test_log.txt
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

if ! [ -z "${release}" ]; then
	if ! [ -z "${linux}" ]; then
		curl -o ecs-cli https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-latest
		curl -o ecs-cli-"${version}" https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v"${version}"
		curl -o ecs-cli-"${hash}" https://s3.amazonaws.com/amazon-ecs-cli/ecs-cli-linux-amd64-v"${hash}"
	fi

	# check that the 3 binaries are the same
	expect_success diff ecs-cli ecs-cli-"${version}"
	expect_success diff ecs-cli ecs-cli-"${hash}"

	chmod +x ecs-cli

	echo "CHECKING VERSION"
	./ecs-cli --version
	if [[ $(./ecs-cli-darwin-amd64-latest --version) == *"${version}"*"${hash}"* ]]; then
		echo -e "${GREEN}SUCCEEDED${NC}: ${LGREEN}ecs-cli --version${NC}"
	else
		echo -e "${RED}FAILED${NC}: ${LRED}ecs-cli --version${NC}"
fi

# configure
expect_success_no_log ./ecs-cli configure --region $region --access-key $access --secret-key $secret --cluster $cluster
# up
expect_success ./ecs-cli up --capability-iam --keypair $keypair --size 2 --instance-type t2.medium --force
# create a service
expect_success ./ecs-cli compose --project-name test2 --file ~/docker-compose.yml service up
# ps on the service
expect_success ./ecs-cli compose --project-name test2 --file ~/docker-compose.yml service ps
# take down service
expect_success ./ecs-cli compose --project-name test2 --file ~/docker-compose.yml service down
# create a task
expect_success ./ecs-cli compose --file ~/docker-compose.yml up
# ps the task
expect_success ./ecs-cli compose --file ~/docker-compose.yml ps
# take down task
expect_success ./ecs-cli compose --file ~/docker-compose.yml down
if [ -z "${release}" ]; then # integration test
	# Service workflow: create, start, scale, stop, down
	expect_success ./ecs-cli compose --project-name test1 --file ~/docker-compose.yml service create
	expect_success ./ecs-cli compose --project-name test1 --file ~/docker-compose.yml service start
	expect_success ./ecs-cli compose --project-name test1 --file ~/docker-compose.yml service scale 2
	expect_success ./ecs-cli compose --project-name test1 --file ~/docker-compose.yml service stop
	expect_success ./ecs-cli compose --project-name test1 --file ~/docker-compose.yml service down
	# task work flow: create, start, scale, then stop
	expect_success ./ecs-cli compose --file ~/docker-compose.yml create
	expect_success ./ecs-cli compose --file ~/docker-compose.yml start
	expect_success ./ecs-cli compose --file ~/docker-compose.yml scale 2
	expect_success ./ecs-cli compose --file ~/docker-compose.yml stop
	# run then stop
	expect_success ./ecs-cli compose --file ~/docker-compose.yml run
	expect_success ./ecs-cli compose --file ~/docker-compose.yml stop
fi


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
