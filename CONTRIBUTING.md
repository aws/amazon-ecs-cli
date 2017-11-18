# Contributing to the CLI
### Setting up your environment
* Make sure you are using go1.8 (`go version`).
* Copy the source code (`go get github.com/aws/amazon-ecs-cli`).

### Building
From `$GOPATH/src/github.com/aws/amazon-ecs-cli`:
* Run `make` (This creates a standalone executable in the `bin/local` directory).

From `$GOPATH/src/github.com/aws/amazon-ecs-cli/ecs-cli`:
* Run `godep restore` (This will download and install dependencies specified in the `Godeps/Godeps.json` into your `$GOPATH`).
* **NOTE:** `godep restore` puts the dependencies in a detached HEAD state (see: [Updating an existing dependency](https://github.com/aws/amazon-ecs-cli/blob/master/README.md#updating-an-existing-dependency)).

### Adding new dependencies
* Make sure you have the latest [godep](https://github.com/tools/godep) (`go get -u github.com/tools/godep`) (version 79)
* `go get` the new dependency.
* Edit your application's source code to import the new dependency.
* From `$GOPATH/src/github.com/aws/amazon-ecs-cli/ecs-cli`, run `godep save ./...` (This will update `Godeps/Godeps.json` and copy the dependencies source to the `vendor/` directory).

### Updating an existing dependency
* While godep recommends using `godep update`, this command unfortunately has several known issues (See: https://github.com/tools/godep/issues/498). We recommend following their [suggestion](https://github.com/tools/godep/issues/498#issuecomment-238946586):
  1. Delete any reference to the package in `Godeps/Godeps.json` (you can also delete the entire `Godeps` dir).
  1. Delete the package from the `/vendor` dir: `rm -rf $GOPATH/src/github.com/aws/amazon-ecs-cli/ecs-cli/vendor/<package>`. (You can also delete the entire vendor dir)
  1. Re-vendor the dependency by running `godep save ./...` from `$GOPATH/src/github.com/aws/amazon-ecs-cli/ecs-cli`.

* Inspect the changes with `git diff` (should show up in `vendor/` directory)
* Example: https://github.com/aws/amazon-ecs-cli/pull/315
* **NOTE:** Unfortunately, using `godep restore` means that `go get` will not work with dependencies. Until we move off `godep`, when we want to update a dependency we will have to go to the dependency in the `$GOPATH` and manually use `git pull` an update to that package.

### Generating mocks
* From `$GOPATH/src/github.com/aws/amazon-ecs-cli/`, run `make generate`. This will look for all files named `generate_mock.go` in the `ecs-cli/modules` directory and call the `mockgen.sh` script, which is a wrapper for the [mockgen](https://github.com/golang/mock#running-mockgen) command.

### Cross-compiling
The `make docker-build` target builds standalone amd64 executables for
Darwin and Linux. The output will be in `bin/darwin-amd64` and `bin/linux-amd64`,
respectively.

If you have set up the appropriate bootstrap environments, you may also directly
run the `make supported-platforms` target to create standalone amd64 executables
for the Darwin and Linux platforms.

### Testing
* To run unit tests, run `make test` from `$GOPATH/src/github.com/aws/amazon-ecs-cli`.

### Licensing
The Amazon ECS CLI is released under an [Apache 2.0](http://aws.amazon.com/apache-2-0/) license. Any code you submit will be released under that license.

For significant changes, we may ask you to sign a [Contributor License Agreement](http://en.wikipedia.org/wiki/Contributor_License_Agreement).


## Amazon Open Source Code of Conduct

This code of conduct provides guidance on participation in Amazon-managed open source communities, and outlines the process for reporting unacceptable behavior. As an organization and community, we are committed to providing an inclusive environment for everyone. Anyone violating this code of conduct may be removed and banned from the community.

**Our open source communities endeavor to:**
* Use welcoming and inclusive language;
* Be respectful of differing viewpoints at all times;
* Accept constructive criticism and work together toward decisions;
* Focus on what is best for the community and users.

**Our Responsibility.** As contributors, members, or bystanders we each individually have the responsibility to behave professionally and respectfully at all times. Disrespectful and unacceptable behaviors include, but are not limited to:
The use of violent threats, abusive, discriminatory, or derogatory language;
* Offensive comments related to gender, gender identity and expression, sexual orientation, disability, mental illness, race, political or religious affiliation;
* Posting of sexually explicit or violent content;
* The use of sexualized language and unwelcome sexual attention or advances;
* Public or private [harassment](http://todogroup.org/opencodeofconduct/#definitions) of any kind;
* Publishing private information, such as physical or electronic address, without permission;
* Other conduct which could reasonably be considered inappropriate in a professional setting;
* Advocating for or encouraging any of the above behaviors.

**Enforcement and Reporting Code of Conduct Issues.**
Instances of abusive, harassing, or otherwise unacceptable behavior may be reported by contacting opensource-codeofconduct@amazon.com. All complaints will be reviewed and investigated and will result in a response that is deemed necessary and appropriate to the circumstances.

**Attribution.** _This code of conduct is based on the [template](http://todogroup.org/opencodeofconduct) established by the [TODO Group](http://todogroup.org/) and the Scope section from the [Contributor Covenant version 1.4](http://contributor-covenant.org/version/1/4/)._
