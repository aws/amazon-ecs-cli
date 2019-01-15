# Amazon ECS CLI - Integration Tests

This directory contains tests intended to run against public Amazon Elastic Container Service endpoints and other AWS Services. They are meant to test end-to-end functionality by executing commands as an ecs-cli user would.

You may be charged for the AWS resources utilized while running these tests. It's not recommended to run these on an AWS account handling production work-loads.

## Test setup

Some tests assume the existance of an "ecs-cli-integ" cluster in the currently configured region. Prior to running these tests, create this cluster using the template in `/resources/ecs_cli_integ_template.json`

## Running the tests

The best way to run them on Linux is via the `make integ-test` target.

The best way to run them on Windows is by building and running the tests with `go` from the project root:
 * `go build -installsuffix cgo -a -ldflags '-s' -o ./bin/local/ecs-cli.exe ./ecs-cli/`
 * `go test -tags integ -v ./ecs-cli/integ/...`