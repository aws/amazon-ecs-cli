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
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

// ecsChunkSize is the maximum number of elements to pass into a describe api
const ecsChunkSize = 100

type ProcessTasksAction func(tasks []*ecs.Task) error

// ECSClient is an interface that specifies only the methods used from the sdk interface. Intended to make mocking and testing easier.
type ECSClient interface {

	// Cluster related
	DeleteCluster(clusterName string) (string, error)
	IsActiveCluster(clusterName string) (bool, error)

	// Service related
	DescribeService(serviceName string) (*ecs.DescribeServicesOutput, error)
	DeleteService(serviceName string) error

	// Task Definition related
	DescribeTaskDefinition(taskDefinitionName string) (*ecs.TaskDefinition, error)

	// Tasks related
	GetTasksPages(listTasksInput *ecs.ListTasksInput, fn ProcessTasksAction) error
	StopTask(taskID string) error
	DescribeTasks(taskIds []*string) ([]*ecs.Task, error)

	// Container Instance related
	GetEC2InstanceIDs(containerInstanceArns []*string) (map[string]string, error)
}

// ecsClient implements ECSClient
type ecsClient struct {
	client ecsiface.ECSAPI
	params *config.CliParams
}

// NewECSClient creates a new ECS client
func NewECSClient(params *config.CliParams) ECSClient {
	ecs := &ecsClient{}
	ecs.Initialize(params)
	return ecs
}

func (c *ecsClient) Initialize(params *config.CliParams) {
	client := ecs.New(params.Session)
	c.client = client
	c.params = params
}

func (c *ecsClient) DeleteCluster(clusterName string) (string, error) {
	resp, err := c.client.DeleteCluster(&ecs.DeleteClusterInput{Cluster: &clusterName})
	if err != nil {
		log.WithFields(log.Fields{
			"cluster": clusterName,
			"error":   err,
		}).Error("Failed to Delete Cluster")
		return "", err
	}
	log.WithFields(log.Fields{
		"cluster": *resp.Cluster.ClusterName,
	}).Info("Deleted cluster")
	return *resp.Cluster.ClusterName, nil
}

func (c *ecsClient) DeleteService(serviceName string) error {
	_, err := c.client.DeleteService(&ecs.DeleteServiceInput{
		Service: aws.String(serviceName),
		Cluster: aws.String(c.params.Cluster),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"service": serviceName,
			"error":   err,
		}).Error("Error deleting service")
		return err
	}
	log.WithFields(log.Fields{"service": serviceName}).Info("Deleted ECS service")
	return nil
}

func (c *ecsClient) DescribeService(serviceName string) (*ecs.DescribeServicesOutput, error) {
	output, err := c.client.DescribeServices(&ecs.DescribeServicesInput{
		Services: []*string{aws.String(serviceName)},
		Cluster:  aws.String(c.params.Cluster),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"service": serviceName,
			"error":   err,
		}).Error("Error describing service")
		return nil, err
	}
	return output, err
}

func (c *ecsClient) DescribeTaskDefinition(taskDefinitionName string) (*ecs.TaskDefinition, error) {
	resp, err := c.client.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	})
	if err != nil {
		return nil, err
	}
	return resp.TaskDefinition, nil

}

// GetTasksPages lists and describe tasks per page and executes the custom function supplied
// any time any call returns error, the processing stops and appropriate error is returned
func (c *ecsClient) GetTasksPages(listTasksInput *ecs.ListTasksInput, tasksFunc ProcessTasksAction) error {
	listTasksInput.Cluster = aws.String(c.params.Cluster)
	var outErr error
	err := c.client.ListTasksPages(listTasksInput, func(page *ecs.ListTasksOutput, end bool) bool {
		if len(page.TaskArns) == 0 {
			return false
		}
		// describe this page of tasks
		resp, err := c.DescribeTasks(page.TaskArns)
		if err != nil {
			outErr = err
			return false
		}
		// execute custom function
		if err = tasksFunc(resp); err != nil {
			outErr = err
			return false
		}
		return true
	})

	if err != nil {
		log.WithFields(log.Fields{
			"request": listTasksInput,
			"error":   err,
		}).Error("Error listing tasks")
		return err
	}
	if outErr != nil {
		return outErr
	}
	return nil
}

func (c *ecsClient) DescribeTasks(taskArns []*string) ([]*ecs.Task, error) {
	descTasksRequest := &ecs.DescribeTasksInput{
		Tasks:   taskArns,
		Cluster: aws.String(c.params.Cluster),
	}
	descTasksResp, err := c.client.DescribeTasks(descTasksRequest)
	if descTasksResp == nil || err != nil {
		log.WithFields(log.Fields{
			"request": descTasksResp,
			"error":   err,
		}).Error("Error describing tasks")
		return nil, err
	}
	return descTasksResp.Tasks, nil
}

func (c *ecsClient) StopTask(taskID string) error {
	_, err := c.client.StopTask(&ecs.StopTaskInput{
		Cluster: aws.String(c.params.Cluster),
		Task:    aws.String(taskID),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"taskId": taskID,
			"error":  err,
		}).Error("Stop task failed")
	}
	return err
}

// GetEC2InstanceIds returns a map of container instance arn to ec2 instance id
func (c *ecsClient) GetEC2InstanceIDs(containerInstanceArns []*string) (map[string]string, error) {
	containerToEC2InstanceMap := map[string]string{}
	for i := 0; i < len(containerInstanceArns); i += ecsChunkSize {
		var chunk []*string
		if i+ecsChunkSize > len(containerInstanceArns) {
			chunk = containerInstanceArns[i:len(containerInstanceArns)]
		} else {
			chunk = containerInstanceArns[i : i+ecsChunkSize]
		}
		descrContainerInstances, err := c.client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(c.params.Cluster),
			ContainerInstances: chunk,
		})
		if err != nil {
			log.WithFields(log.Fields{
				"containerInstancesCount": len(containerInstanceArns),
				"error":                   err,
			}).Error("Error describing container instance")
			return nil, err
		}
		for _, containerInstance := range descrContainerInstances.ContainerInstances {
			if containerInstance.Ec2InstanceId != nil {
				containerToEC2InstanceMap[aws.StringValue(containerInstance.ContainerInstanceArn)] = aws.StringValue(containerInstance.Ec2InstanceId)
			}
		}
	}
	return containerToEC2InstanceMap, nil
}

// IsActiveCluster returns true if the cluster exists and can be described.
func (c *ecsClient) IsActiveCluster(clusterName string) (bool, error) {
	output, err := c.client.DescribeClusters(&ecs.DescribeClustersInput{
		Clusters: []*string{aws.String(clusterName)},
	})

	if err != nil {
		return false, err
	}

	if len(output.Failures) > 0 {
		return false, nil
	} else if len(output.Clusters) == 0 {
		return false, fmt.Errorf("Got an empty list of clusters while describing the cluster '%s'", clusterName)
	}

	status := aws.StringValue(output.Clusters[0].Status)
	if "ACTIVE" == status {
		return true, nil
	}

	log.WithFields(log.Fields{"cluster": clusterName, "status": status}).Debug("cluster status")
	return false, nil
}
