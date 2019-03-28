package cmd_up

import (
	"fmt"
	"github.com/aws/amazon-ecs-cli/ecs-cli/integ"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCusterCreationWithGPUInstances(t *testing.T) {
	// Given
	cfnClient, ecsClient, clusterName := setup(t)
	cmd := integ.GetCommand([]string{"up", "-c", clusterName, "--instance-type", "p2.xlarge", "--capability-iam", "--force"})

	// When
	stdout, err := cmd.Output()
	assert.NoError(t, err, fmt.Sprintf("Error running %v\nStdout: %s", cmd.Args, string(stdout)))

	// Then
	assertHasCFNStack(t, cfnClient, clusterName)
	assertHasActiveContainerInstances(t, ecsClient, clusterName, 1) // by default we only activate 1 instance
	assertHasGPUResources(t, ecsClient, clusterName)


	// Cleanup the created resources
	after(cfnClient, ecsClient, clusterName)
}

func assertHasGPUResources(t *testing.T, client *ecs.ECS, clusterName string) {
	cluster, err := client.ListContainerInstances(&ecs.ListContainerInstancesInput{
		Cluster: aws.String(clusterName),
	})

	if err != nil {
		assert.FailNow(t, "failed to list container instances in the cluster", clusterName, err.Error())
	}
	instances, err := client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
		Cluster:            aws.String(clusterName),
		ContainerInstances: cluster.ContainerInstanceArns,
	})
	if err != nil {
		assert.FailNow(t, "failed to describe container instances in the cluster", clusterName, err.Error())
	}

	for _, instance := range instances.ContainerInstances {
		hasGPU := false
		for _, resource := range instance.RegisteredResources {
			hasGPU = hasGPU || *resource.Name == "GPU"
		}
		assert.True(t, hasGPU, "Expected instance %s to have GPU resource", *instance.Ec2InstanceId)
	}
}