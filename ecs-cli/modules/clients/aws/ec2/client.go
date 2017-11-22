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

package ec2

import (
	"errors"

	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/clients"
	"github.com/aws/amazon-ecs-cli/ecs-cli/modules/config"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/ec2"
	"github.com/aws/aws-sdk-go/service/ec2/ec2iface"
)

// EC2Client defines methods to interact with the EC2 API interface.
type EC2Client interface {
	DescribeInstances(ec2InstanceIds []*string) (map[string]*ec2.Instance, error)
	DescribeNetworkInterfaces(networkInterfaceIDs []*string) ([]*ec2.NetworkInterface, error)
}

// ec2Client implements EC2Client
type ec2Client struct {
	client ec2iface.EC2API
}

// NewEC2Client creates an instance of ec2Client object.
func NewEC2Client(params *config.CLIParams) EC2Client {
	client := ec2.New(params.Session)
	client.Handlers.Build.PushBackNamed(clients.CustomUserAgentHandler())
	return newClient(params, client)
}

func newClient(params *config.CLIParams, client ec2iface.EC2API) EC2Client {
	return &ec2Client{
		client: client,
	}
}

// DescribeInstances returns a map of instanceId to EC2 Instance
func (c *ec2Client) DescribeInstances(ec2InstanceIds []*string) (map[string]*ec2.Instance, error) {
	if len(ec2InstanceIds) == 0 {
		return make(map[string]*ec2.Instance, 0), nil
	}
	output, err := c.client.DescribeInstances(&ec2.DescribeInstancesInput{
		InstanceIds: ec2InstanceIds,
	})
	if err != nil {
		return nil, err
	}

	ec2Instances := map[string]*ec2.Instance{}
	if output.Reservations == nil || len(output.Reservations) == 0 {
		return nil, errors.New("No EC2 reservations found")
	}
	for _, reservation := range output.Reservations {
		for _, ec2Instance := range reservation.Instances {
			if ec2Instance.InstanceId == nil {
				continue
			}
			ec2Instances[aws.StringValue(ec2Instance.InstanceId)] = ec2Instance
		}
	}
	return ec2Instances, nil
}

func (c *ec2Client) DescribeNetworkInterfaces(networkInterfaceIDs []*string) ([]*ec2.NetworkInterface, error) {
	request := &ec2.DescribeNetworkInterfacesInput{
		NetworkInterfaceIds: networkInterfaceIDs,
	}
	response, err := c.client.DescribeNetworkInterfaces(request)
	if err != nil {
		return nil, err
	}
	return response.NetworkInterfaces, nil
}
