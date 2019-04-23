// Copyright 2015-2019 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//	http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package cmd

import (
	"fmt"
	"testing"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"

	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/cfn"
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ/stdout"

	"github.com/stretchr/testify/assert"
)

// TestDown runs `ecs-cli down` to remove a cluster.
func TestDown(t *testing.T, conf *CLIConfig) {
	// Given
	args := []string{
		"down",
		"--force",
		"--cluster-config",
		conf.ConfigName,
	}
	cmd := integ.GetCommand(args)

	// When
	out, err := cmd.Output()
	if err != nil {
		assert.FailNowf(t, "Failed to delete cluster", "Error %v running %v", err, args)
	}

	// Then
	stdout.Stdout(out).TestHasAllSnippets(t, []string{
		"Deleted cluster",
		fmt.Sprintf("cluster=%s", conf.ClusterName),
	})
	cfn.TestNoStackName(t, stackName(conf.ClusterName))
	t.Logf("Deleted stack %s", stackName(conf.ClusterName))
}
