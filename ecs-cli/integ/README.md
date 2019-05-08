# Amazon ECS CLI - Integration Tests

This directory contains tests intended to run against public Amazon Elastic Container Service endpoints and other AWS Services. They are meant to test end-to-end functionality by executing commands as an ecs-cli user would.

You may be charged for the AWS resources utilized while running these tests. It's not recommended to run these on an AWS account handling production work-loads.

## Local test setup

1. Set the following mandatory environment variables:
    ```bash
    # hardcode the CodeBuild ID since we're not using codebuild to run the tests
    export CODEBUILD_BUILD_ID="local-integ-test"
    # change this to the region you'd like to run the tests in
    export AWS_REGION="us-east-1"
    ```
2. If your OS doesn't support the `TMPDIR` environment variable, then set it:
    ```bash
    # If your OS sets TMPDIR, you should get an output like /var/folders/13/y9bcvw7557d5bvlvrj8jz0k04gs5dl/T/
    echo $TMPDIR
    # Otherwise set the environment to a location of your choice
    export $TMPDIR="/tmp"
    ```   


## Running the tests

The best way to run them on Linux is via the `make integ-test` target.

The best way to run them on Windows is by building and running the tests with `go` from the project root:
 * `go build -installsuffix cgo -a -ldflags '-s' -o ./bin/local/ecs-cli.exe ./ecs-cli/`
 * `go test -tags integ -v ./ecs-cli/integ/...`