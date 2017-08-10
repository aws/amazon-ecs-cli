<!--
 Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.

 Licensed under the Apache License, Version 2.0 (the "License"). You may
 not use this file except in compliance with the License. A copy of the
 License is located at

 http://aws.amazon.com/apache2.0/

 or in the "license" file accompanying this file. This file is distributed
 on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
 express or implied. See the License for the specific language governing
 permissions and limitations under the License.
-->

# Introduction

Currently the CLI only has unit tests for each package, there is no way to programmatically ensure that commands work as intended. This requires the team to waste time manually testing and running commands, and increases the likelihood of buggy code being released to the public (which decreases customer trust in our CLI).

# Proposal

## Tooling

Many tools exist for testing command line tools, however, in general their purpose is to verify that the terminal output of a command is correct (Example https://en.wikipedia.org/wiki/Expect). For the ECS CLI, this should not be the primary goal- the output of many of our commands is not deterministic (for example, the output is often a stream of updates on the status of a cluster/service/task/etc). Moreover, the output that is most important to test from a user point of view, is not the terminal text output, but the affects of the command caused on AWS services. For example, our primary aim in testing the cluster up command is that a cluster is actually created in ECS with the correct number and type of instances, etc. (This is not to suggest however that output verification is not something that we would eventually want to have as well).

Therefore, the AWS go SDK will be used to make calls to AWS to verify that the resources are correctly being created. The integration tests can be written in go, and committed to the existing ECS CLI repository.

## Workflow

Here is the flow for how the unit tests will work:
1. TestMain function sets up AWS resources needed for the test
  * Key Pairs
  * Default clusters to deploy tasks/services to
  * Configure the CLI
2. Individual test cases will test commands
  1. Use the `os` library to shell out command
  2. Use the AWS SDK to verify that the correct affects on AWS resources have occurred (these checks could be deep or shallow, depending upon how much effort we want to put into these tests).


## Demo

This proposal is accompanied with a demo, which contains code to run a single very simple test case. This can be quickly built upon to create new test cases. It uses `go test`, the built in testing framework for Golang. The tests use testing.Short to ensure that IDEs like Atom won't automatically run the integration tests.
(The demo is currently just skeleton code- 0.5-1 days of dedicated effort is needed to complete it.)

## Plan

At this point in time, no one on the team has enough bandwidth to dedicate the time necessary to make the unit tests a reality. Here is a more realistic plan for how we could add integration tests to the whole CLI:
- 1-2 days more time can be spent to set things up and add a few basic test cases
- Mandate that for every new pull request/feature, there must be new integration tests- over time this will allow us to slowly build up integration tests overtime, without slowing down the release of any individual feature much at all.

## Design Concerns

The main issue with any proposal for automated running of commands is that an AWS account will be needed for testing, and the tests themselves need to have credentials. In this proposal, this has been solved using environment variables to specify the credentials for an account. However, it is uncertain whether this is the best option- comment and counter proposals are welcomed. 

## Future Work

Eventually, the integration tests can be integrated with Travis and run automatically when someone creates a PR on Github.
