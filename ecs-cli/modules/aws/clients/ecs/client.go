// Copyright 2015-2016 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

package ecs

import (
	"crypto/md5"
	"errors"
	"fmt"

	log "github.com/Sirupsen/logrus"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/compose/ecs/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/amazon-ecs-cli/ecs-cli/utils/cache"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/aws/aws-sdk-go/service/ecs/ecsiface"
)

//go:generate mockgen.sh github.com/aws/amazon-ecs-cli/ecs-cli/modules/aws/clients/ecs ECSClient mock/$GOFILE
//go:generate mockgen.sh github.com/aws/aws-sdk-go/service/ecs/ecsiface ECSAPI mock/sdk/ecsiface_mock.go

// ecsChunkSize is the maximum number of elements to pass into a describe api
const ecsChunkSize = 100

type ProcessTasksAction func(tasks []*ecs.Task) error

// ECSClient is an interface that specifies only the methods used from the sdk interface. Intended to make mocking and testing easier.
type ECSClient interface {
	// TODO: Modify the interface and tbe client to not have the Initialize method.
	Initialize(params *config.CliParams)

	// Cluster related
	CreateCluster(clusterName string) (string, error)
	DeleteCluster(clusterName string) (string, error)
	IsActiveCluster(clusterName string) (bool, error)

	// Service related
	CreateService(serviceName, taskDefName string, deploymentConfig *ecs.DeploymentConfiguration) error
	UpdateServiceCount(serviceName string, count int64, deploymentConfig *ecs.DeploymentConfiguration) error
	UpdateService(serviceName, taskDefinitionName string, count int64, deploymentConfig *ecs.DeploymentConfiguration) error
	DescribeService(serviceName string) (*ecs.DescribeServicesOutput, error)
	DeleteService(serviceName string) error

	// Task Definition related
	RegisterTaskDefinition(request *ecs.RegisterTaskDefinitionInput) (*ecs.TaskDefinition, error)
	RegisterTaskDefinitionIfNeeded(request *ecs.RegisterTaskDefinitionInput, tdCache cache.Cache) (*ecs.TaskDefinition, error)
	DescribeTaskDefinition(taskDefinitionName string) (*ecs.TaskDefinition, error)

	// Tasks related
	GetTasksPages(listTasksInput *ecs.ListTasksInput, fn ProcessTasksAction) error
	RunTask(taskDefinition, startedBy string, count int) (*ecs.RunTaskOutput, error)
	RunTaskWithOverrides(taskDefinition, startedBy string, count int, overrides map[string]string) (*ecs.RunTaskOutput, error)
	StopTask(taskId string) error
	DescribeTasks(taskIds []*string) ([]*ecs.Task, error)

	// Container Instance related
	GetEC2InstanceIDs(containerInstanceArns []*string) (map[string]string, error)
}

// ecsClient implements ECSClient
type ecsClient struct {
	client ecsiface.ECSAPI
	params *config.CliParams
}

func NewECSClient() ECSClient {
	return &ecsClient{}
}

func (c *ecsClient) Initialize(params *config.CliParams) {
	client := ecs.New(session.New(params.Config))
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	c.client = client
	c.params = params
}

func (c *ecsClient) CreateCluster(clusterName string) (string, error) {
	resp, err := c.client.CreateCluster(&ecs.CreateClusterInput{ClusterName: &clusterName})
	if err != nil {
		log.WithFields(log.Fields{
			"cluster": clusterName,
			"error":   err,
		}).Error("Failed to Create Cluster")
		return "", err
	}
	log.WithFields(log.Fields{
		"cluster": *resp.Cluster.ClusterName,
	}).Info("Created cluster")

	return *resp.Cluster.ClusterName, nil
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

func (client *ecsClient) DeleteService(serviceName string) error {
	_, err := client.client.DeleteService(&ecs.DeleteServiceInput{
		Service: aws.String(serviceName),
		Cluster: aws.String(client.params.Cluster),
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

func (client *ecsClient) CreateService(serviceName, taskDefName string, deploymentConfig *ecs.DeploymentConfiguration) error {
	_, err := client.client.CreateService(&ecs.CreateServiceInput{
		DesiredCount:            aws.Int64(0),            // Required
		ServiceName:             aws.String(serviceName), // Required
		TaskDefinition:          aws.String(taskDefName), // Required
		Cluster:                 aws.String(client.params.Cluster),
		DeploymentConfiguration: deploymentConfig,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"service": serviceName,
			"error":   err,
		}).Error("Error creating service")
		return err
	}

	fields := log.Fields{
		"service":        serviceName,
		"taskDefinition": taskDefName,
	}
	if deploymentConfig != nil && deploymentConfig.MaximumPercent != nil {
		fields["deployment-max-percent"] = aws.Int64Value(deploymentConfig.MaximumPercent)
	}
	if deploymentConfig != nil && deploymentConfig.MinimumHealthyPercent != nil {
		fields["deployment-min-healthy-percent"] = aws.Int64Value(deploymentConfig.MinimumHealthyPercent)
	}

	log.WithFields(fields).Info("Created an ECS service")
	return nil
}

func (client *ecsClient) UpdateServiceCount(serviceName string, count int64, deploymentConfig *ecs.DeploymentConfiguration) error {
	return client.UpdateService(serviceName, "", count, deploymentConfig)
}

func (client *ecsClient) UpdateService(serviceName, taskDefinition string, count int64, deploymentConfig *ecs.DeploymentConfiguration) error {
	input := &ecs.UpdateServiceInput{
		DesiredCount:            aws.Int64(count),
		Service:                 aws.String(serviceName),
		Cluster:                 aws.String(client.params.Cluster),
		DeploymentConfiguration: deploymentConfig,
	}
	if taskDefinition != "" {
		input.TaskDefinition = aws.String(taskDefinition)
	}
	_, err := client.client.UpdateService(input)
	if err != nil {
		log.WithFields(log.Fields{
			"service": serviceName,
			"error":   err,
		}).Error("Error updating service")
		return err
	}
	fields := log.Fields{
		"service": serviceName,
		"count":   count,
	}
	if taskDefinition != "" {
		fields["taskDefinition"] = taskDefinition
	}
	if deploymentConfig != nil && deploymentConfig.MaximumPercent != nil {
		fields["deployment-max-percent"] = aws.Int64Value(deploymentConfig.MaximumPercent)
	}
	if deploymentConfig != nil && deploymentConfig.MinimumHealthyPercent != nil {
		fields["deployment-min-healthy-percent"] = aws.Int64Value(deploymentConfig.MinimumHealthyPercent)
	}
	log.WithFields(fields).Debug("Updated ECS service")
	return nil
}

func (client *ecsClient) DescribeService(serviceName string) (*ecs.DescribeServicesOutput, error) {
	output, err := client.client.DescribeServices(&ecs.DescribeServicesInput{
		Services: []*string{aws.String(serviceName)},
		Cluster:  aws.String(client.params.Cluster),
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

func (client *ecsClient) RegisterTaskDefinition(request *ecs.RegisterTaskDefinitionInput) (*ecs.TaskDefinition, error) {
	resp, err := client.client.RegisterTaskDefinition(request)
	if err != nil {
		log.WithFields(log.Fields{
			"family": aws.StringValue(request.Family),
			"error":  err,
		}).Error("Error registering task definition")
		return nil, err
	}
	return resp.TaskDefinition, nil
}

// RegisterTaskDefinitionIfNeeded checks if a task definition has already been
// registered via the provided cache, and if so returns it.
// Otherwise, it registers a new one.
//
// This exists to avoid an explosion of task definitions for automatically
// registered inputs.
func (client *ecsClient) RegisterTaskDefinitionIfNeeded(
	request *ecs.RegisterTaskDefinitionInput,
	taskDefinitionCache cache.Cache) (*ecs.TaskDefinition, error) {
	if request.Family == nil {
		return nil, errors.New("invalid task definitions: family is required")
	}

	taskDefResp, err := client.DescribeTaskDefinition(aws.StringValue(request.Family))

	// If there are no task definitions for this family OR the task definition exists and is marked as 'INACTIVE',
	// register the task definition and create a cache entry
	if err != nil || *taskDefResp.Status == ecs.TaskDefinitionStatusInactive {
		return persistTaskDefinition(request, client, taskDefinitionCache)
	}

	tdHash := client.constructTaskDefinitionCacheHash(taskDefResp, request)

	td := &ecs.TaskDefinition{}
	err = taskDefinitionCache.Get(tdHash, td)
	if err != nil || !cachedTaskDefinitionRevisionIsActive(td, client) {
		log.WithFields(log.Fields{
			"taskDefHash": tdHash,
			"taskDef":     td,
		}).Debug("cache miss")
		return persistTaskDefinition(request, client, taskDefinitionCache)
	}

	log.WithFields(log.Fields{
		"taskDefHash": tdHash,
		"taskDef":     td,
	}).Debug("cache hit")
	return td, nil
}

// cachedTaskDefinitionRevisionIsActive asserts that the family:revison for both the locally cached Task Definition and the Task Definition stored in ECS is listed as ACTIVE
func cachedTaskDefinitionRevisionIsActive(cachedTaskDefinition *ecs.TaskDefinition, client *ecsClient) bool {
	taskDefinitionOfRecord, err := client.DescribeTaskDefinition(aws.StringValue(cachedTaskDefinition.TaskDefinitionArn))
	if err != nil || taskDefinitionOfRecord == nil {
		log.WithFields(log.Fields{
			"taskDefinitionName": aws.StringValue(cachedTaskDefinition.TaskDefinitionArn),
			"error":              err,
		}).Error("Error describing task definition")
		return false
	}
	return *taskDefinitionOfRecord.Status == ecs.TaskDefinitionStatusActive
}

// constructTaskDefinitionCacheHash computes md5sum of the region, awsAccountId and the requested task definition data
// BUG(juanrhenals) The requested Task Definition data (taskDefinitionRequest) is not created in a deterministic fashion because there are maps within
// the request ecs.RegisterTaskDefinitionInput structure, and map iteration in Go is not deterministic. We need to fix this.
func (client *ecsClient) constructTaskDefinitionCacheHash(taskDefinition *ecs.TaskDefinition, request *ecs.RegisterTaskDefinitionInput) string {
	// Get the region from the ecsClient configuration
	region := aws.StringValue(client.params.Config.Region)
	awsUserAccountId := utils.GetAwsAccountIdFromArn(aws.StringValue(taskDefinition.TaskDefinitionArn))
	tdHashInput := fmt.Sprintf("%s-%s-%s", region, awsUserAccountId, request.GoString())
	return fmt.Sprintf("%x", md5.Sum([]byte(tdHashInput)))
}

// persistTaskDefinition registers the task definition with ECS and creates a new local cache entry
func persistTaskDefinition(request *ecs.RegisterTaskDefinitionInput, client *ecsClient, taskDefinitionCache cache.Cache) (*ecs.TaskDefinition, error) {
	resp, err := client.RegisterTaskDefinition(request)
	if err != nil {
		return nil, err
	}

	tdHash := client.constructTaskDefinitionCacheHash(resp, request)

	err = taskDefinitionCache.Put(tdHash, resp)
	if err != nil {
		log.WithFields(log.Fields{
			"error": err,
		}).Warn("Could not cache task definition; redundant task definitions might be created")
		// We can keep going even if we can't cache and operate mostly fine
	}
	return resp, err

}

func (client *ecsClient) DescribeTaskDefinition(taskDefinitionName string) (*ecs.TaskDefinition, error) {
	resp, err := client.client.DescribeTaskDefinition(&ecs.DescribeTaskDefinitionInput{
		TaskDefinition: aws.String(taskDefinitionName),
	})
	if err != nil {
		return nil, err
	}
	return resp.TaskDefinition, nil

}

// GetTasksPages lists and describe tasks per page and executes the custom function supplied
// any time any call returns error, the processing stops and appropriate error is returned
func (client *ecsClient) GetTasksPages(listTasksInput *ecs.ListTasksInput, tasksFunc ProcessTasksAction) error {
	listTasksInput.Cluster = aws.String(client.params.Cluster)
	var outErr error
	err := client.client.ListTasksPages(listTasksInput, func(page *ecs.ListTasksOutput, end bool) bool {
		if len(page.TaskArns) == 0 {
			return false
		}
		// describe this page of tasks
		resp, err := client.DescribeTasks(page.TaskArns)
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

func (client *ecsClient) DescribeTasks(taskArns []*string) ([]*ecs.Task, error) {
	descTasksRequest := &ecs.DescribeTasksInput{
		Tasks:   taskArns,
		Cluster: aws.String(client.params.Cluster),
	}
	descTasksResp, err := client.client.DescribeTasks(descTasksRequest)
	if descTasksResp == nil || err != nil {
		log.WithFields(log.Fields{
			"request": descTasksResp,
			"error":   err,
		}).Error("Error describing tasks")
		return nil, err
	}
	return descTasksResp.Tasks, nil
}

// RunTask issues a run task request for the input task definition
func (client *ecsClient) RunTask(taskDefinition, startedBy string, count int) (*ecs.RunTaskOutput, error) {
	resp, err := client.client.RunTask(&ecs.RunTaskInput{
		Cluster:        aws.String(client.params.Cluster),
		TaskDefinition: aws.String(taskDefinition),
		StartedBy:      aws.String(startedBy),
		Count:          aws.Int64(int64(count)),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"task definition": taskDefinition,
			"error":           err,
		}).Error("Error running tasks")
	}
	return resp, err
}

// RunTask issues a run task request for the input task definition
func (client *ecsClient) RunTaskWithOverrides(taskDefinition, startedBy string, count int, overrides map[string]string) (*ecs.RunTaskOutput, error) {

	commandOverrides := []*ecs.ContainerOverride{}
	for cont, command := range overrides {
		contOverride := &ecs.ContainerOverride{
			Name:    aws.String(cont),
			Command: aws.StringSlice([]string{command}),
		}
		commandOverrides = append(commandOverrides, contOverride)
	}
	ecsOverrides := &ecs.TaskOverride{
		ContainerOverrides: commandOverrides,
	}

	resp, err := client.client.RunTask(&ecs.RunTaskInput{
		Cluster:        aws.String(client.params.Cluster),
		TaskDefinition: aws.String(taskDefinition),
		StartedBy:      aws.String(startedBy),
		Count:          aws.Int64(int64(count)),
		Overrides:      ecsOverrides,
	})
	if err != nil {
		log.WithFields(log.Fields{
			"task definition": taskDefinition,
			"error":           err,
		}).Error("Error running tasks")
	}
	return resp, err
}

func (client *ecsClient) StopTask(taskId string) error {
	_, err := client.client.StopTask(&ecs.StopTaskInput{
		Cluster: aws.String(client.params.Cluster),
		Task:    aws.String(taskId),
	})
	if err != nil {
		log.WithFields(log.Fields{
			"taskId": taskId,
			"error":  err,
		}).Error("Stop task failed")
	}
	return err
}

// GetEC2InstanceIds returns a map of container instance arn to ec2 instance id
func (client *ecsClient) GetEC2InstanceIDs(containerInstanceArns []*string) (map[string]string, error) {
	containerToEC2InstanceMap := map[string]string{}
	for i := 0; i < len(containerInstanceArns); i += ecsChunkSize {
		var chunk []*string
		if i+ecsChunkSize > len(containerInstanceArns) {
			chunk = containerInstanceArns[i:len(containerInstanceArns)]
		} else {
			chunk = containerInstanceArns[i : i+ecsChunkSize]
		}
		descrContainerInstances, err := client.client.DescribeContainerInstances(&ecs.DescribeContainerInstancesInput{
			Cluster:            aws.String(client.params.Cluster),
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
func (client *ecsClient) IsActiveCluster(clusterName string) (bool, error) {
	output, err := client.client.DescribeClusters(&ecs.DescribeClustersInput{
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
