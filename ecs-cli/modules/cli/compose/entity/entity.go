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

package entity

import (
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/context"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/cli/compose/entity/types"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/utils/cache"
	"github.com/aws/aws-sdk-go/service/ecs"
	"github.com/docker/libcompose/project"
)

// ProjectEntity ties closely to how operations performed with the compose yaml are integrated with ECS
// It holds all the commands that are needed to operate the compose app
type ProjectEntity interface {
	Create() error
	Start() error
	Up() error
	Info(filterComposeTasks bool) (project.InfoSet, error)
	Run(commandOverrides map[string][]string) error
	Scale(count int) error
	Stop() error
	Down() error

	LoadContext() error
	Context() *context.ECSContext
	Sleeper() *utils.TimeSleeper
	TaskDefinition() *ecs.TaskDefinition
	TaskDefinitionCache() cache.Cache
	SetTaskDefinition(taskDefinition *ecs.TaskDefinition)
	EntityType() types.Type
}
