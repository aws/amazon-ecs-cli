// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package integration

import (
	"os"
	"os/exec"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
)

const (
	TestClusterPrefix = "Integration-Test-Main-Cluster-"
)

func TestMain(m *testing.M) {
	if checkEnv() {
		setup()
		code := m.Run()
		os.Exit(code)
	} else {
		logrus.Error("Environment Not Set")
		os.Exit(1)
	}
}

func GetDefaultClusterName() string {
	return TestClusterPrefix + time.Now().UTC().Format("01-02--03_04_05PM")
}

// do setup for the integration tests
func setup() error {
	// Configure the cli
	region := os.Getenv("INTEGRATION_TEST_PRIMARY_REGION")
	accessKey := os.Getenv("INTEGRATION_TEST_ACCESS_KEY")
	secretKey := os.Getenv("INTEGRATION_TEST_SECRET_KEY")
	if err := ConfigureCommand(GetDefaultClusterName(), region, accessKey, secretKey); err != nil {
		return err
	}

	return nil
}

func ConfigureCommand(cluster string, region string, accessKey string, secretKey string) error {
	// ecs-cli configure --region $INTEGRATION_TEST_PRIMARY_REGION --access-key $INTEGRATION_TEST_ACCESS_KEY --secret-key $INTEGRATION_TEST_SECRET_KEY --cluster defualt
	cmdArgs := []string{"configure", "--region", region, "--access-key", accessKey, "--secret-key", secretKey, "--cluster", cluster}
	if _, err := exec.Command("ecs-cli", cmdArgs...).Output(); err != nil {
		return err
	}

	return nil
}

func ClusterUpCommand(cluster string, region string, accessKey string, secretKey string) error {
	// ecs-cli configure --region $INTEGRATION_TEST_PRIMARY_REGION --access-key $INTEGRATION_TEST_ACCESS_KEY --secret-key $INTEGRATION_TEST_SECRET_KEY --cluster defualt
	cmdArgs := []string{"configure", "--region", region, "--access-key", accessKey, "--secret-key", secretKey, "--cluster", cluster}
	if _, err := exec.Command("ecs-cli", cmdArgs...).Output(); err != nil {
		return err
	}

	return nil
}

func CreateEcsClient() (ECSClient, error) {
	ecsConfig := &config.CliConfig{}
	ecsConfig.Cluster = GetDefaultClusterName()
	ecsConfig.Region = os.Getenv("INTEGRATION_TEST_PRIMARY_REGION")
	ecsConfig.AwsAccessKey = os.Getenv("INTEGRATION_TEST_ACCESS_KEY")
	ecsConfig.AwsSecretKey = os.Getenv("INTEGRATION_TEST_SECRET_KEY")

	svcSession, err := ecsConfig.ToAWSSession()
	if err != nil {
		return nil, err
	}

	params := &config.CliParams{
		Cluster: ecsConfig.Cluster,
		Session: svcSession,
	}

	return NewECSClient(params), nil
}

func LongRunningTest(t *testing.T) {
	if testing.Short() {
		// prevents test from running when -short flag is used
		// Prevents Atom from auto-running the test
		t.Skip("skipping test in short mode.")
	}
}

func checkEnv() bool {
	if os.Getenv("INTEGRATION_TEST_ACCESS_KEY") == "" {
		return false
	}
	if os.Getenv("INTEGRATION_TEST_SECRET_KEY") == "" {
		return false
	}
	if os.Getenv("INTEGRATION_TEST_PRIMARY_REGION") == "" {
		return false
	}
	if os.Getenv("INTEGRATION_TEST_SECONDARY_REGION") == "" {
		return false
	}

	return true
}
