// Copyright 2015-2018 Amazon.com, Inc. or its affiliates. All Rights Reserved.
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

// Package servicediscovery contains functions for working with the route53 APIs
// that back ECS Service Discovery
package route53

import (
	"testing"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/service/route53"
	"github.com/aws/aws-sdk-go/service/servicediscovery"
	"github.com/stretchr/testify/assert"
)

func TestFindPrivateNamespace(t *testing.T) {
	mockSD := setupNamespaceMocks(t, false)
	mockR53 := setupHostedZoneMocks(t)

	var testCases = []struct {
		testName   string
		name       string
		vpc        string
		region     string
		expectedID *string
	}{
		{
			testName:   "Find namespace1",
			name:       "corp",
			vpc:        "vpc-8BAADF00D",
			region:     "us-east-1",
			expectedID: aws.String("namespace1"),
		},
		{
			testName:   "Find namespace2",
			name:       "prod",
			vpc:        "vpc-1CEB00DA",
			region:     "us-east-1",
			expectedID: aws.String("namespace2"),
		},
		{
			testName:   "Find namespace4",
			name:       "corp",
			vpc:        "vpc-C00010FF",
			region:     "sa-east-1",
			expectedID: aws.String("namespace4"),
		},
		{
			testName:   "Namespace Does Not Exist",
			name:       "corp",
			vpc:        "vpc-C00010FF",
			region:     "us-east-1",
			expectedID: nil,
		},
		{
			testName:   "Namspace with multiple VPCs",
			name:       "bridge",
			vpc:        "vpc-D15EA5E",
			region:     "ap-south-1",
			expectedID: aws.String("namespace3"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			namespaceID, err := findPrivateNamespace(testCase.name, testCase.vpc, testCase.region, mockR53, mockSD)
			assert.Equal(t, testCase.expectedID, namespaceID, "Expected namespace ID to match")
			assert.NoError(t, err, "Unexpected error from FindPrivateNamespace")
		})
	}
}

func TestFindPublicNamespace(t *testing.T) {
	mockSD := setupNamespaceMocks(t, true)

	var testCases = []struct {
		testName   string
		name       string
		expectedID *string
	}{
		{
			testName:   "Find namespace1",
			name:       "corp",
			expectedID: aws.String("namespace1"),
		},
		{
			testName:   "Find namespace2",
			name:       "prod",
			expectedID: aws.String("namespace2"),
		},
		{
			testName:   "Find namespace3",
			name:       "bridge",
			expectedID: aws.String("namespace3"),
		},
		{
			testName:   "Namespace Does Not Exist",
			name:       "cat",
			expectedID: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.testName, func(t *testing.T) {
			namespaceID, err := findPublicNamespace(testCase.name, mockSD)
			assert.Equal(t, testCase.expectedID, namespaceID, "Expected namespace ID to match")
			assert.NoError(t, err, "Unexpected error from FindPrivateNamespace")
		})
	}
}

func TestWaitUntilSDSDeletable(t *testing.T) {
	mockSD := &mockSDClient{
		sdsInstanceCount: 3,
	}

	err := waitUntilSDSDeletable("someid", mockSD, 5)
	assert.NoError(t, err, "Unexpected error calling waitUntilSDSDeletable")
	assert.Equal(t, 4, mockSD.getServiceCallCount, "Expected GetService() to be called 4 times")
}

func TestWaitUntilSDSDeletableErrorCase(t *testing.T) {
	mockSD := &mockSDClient{
		sdsInstanceCount: 6,
	}

	err := waitUntilSDSDeletable("someid", mockSD, 5)
	assert.Error(t, err, "Expected error calling waitUntilSDSDeletable")
	assert.Equal(t, 5, mockSD.getServiceCallCount, "Expected GetService() to be called 5 times")
}

// Implements serviceDiscoveryClient interface
type mockSDClient struct {
	usingFilter      bool
	publicNamespaces bool
	namespaceData    map[string]servicediscovery.Namespace
	t                *testing.T
	// value is decremented for each call to GetService, this is used when testing waitUntilSDSDeletable
	sdsInstanceCount int64
	// count of calls to GetService
	getServiceCallCount int
}

func (mock *mockSDClient) ListNamespacesPages(input *servicediscovery.ListNamespacesInput, fn func(*servicediscovery.ListNamespacesOutput, bool) bool) error {
	if mock.usingFilter {
		filter := aws.StringValue(input.Filters[0].Values[0])
		if mock.publicNamespaces {
			assert.Equal(mock.t, servicediscovery.NamespaceTypeDnsPublic, filter)
		} else {
			assert.Equal(mock.t, servicediscovery.NamespaceTypeDnsPrivate, filter)
		}
	}

	var namespaces []*servicediscovery.NamespaceSummary
	for _, namespace := range mock.namespaceData {
		namespaceSummary := servicediscovery.NamespaceSummary{
			Id:   namespace.Id,
			Name: namespace.Name,
		}
		namespaces = append(namespaces, &namespaceSummary)
	}
	apiOutput := &servicediscovery.ListNamespacesOutput{
		Namespaces: namespaces,
	}

	fn(apiOutput, true)
	return nil
}

func (mock *mockSDClient) GetNamespace(input *servicediscovery.GetNamespaceInput) (*servicediscovery.GetNamespaceOutput, error) {
	namespace := mock.namespaceData[aws.StringValue(input.Id)]
	return &servicediscovery.GetNamespaceOutput{
		Namespace: &namespace,
	}, nil
}

func (mock *mockSDClient) GetService(input *servicediscovery.GetServiceInput) (*servicediscovery.GetServiceOutput, error) {
	currentCount := mock.sdsInstanceCount
	mock.sdsInstanceCount--
	mock.getServiceCallCount++
	return &servicediscovery.GetServiceOutput{
		Service: &servicediscovery.Service{
			InstanceCount: aws.Int64(currentCount),
		},
	}, nil
}

// Implements route53Client interface
type mockRoute53Client struct {
	hostedZoneData map[string]route53.GetHostedZoneOutput
}

func (mock *mockRoute53Client) GetHostedZone(input *route53.GetHostedZoneInput) (*route53.GetHostedZoneOutput, error) {
	zone := mock.hostedZoneData[aws.StringValue(input.Id)]
	return &zone, nil
}

func setupHostedZoneMocks(t *testing.T) route53Client {
	// Mock Data
	var hostedZoneData = map[string]route53.GetHostedZoneOutput{
		"zone1": route53.GetHostedZoneOutput{
			VPCs: []*route53.VPC{
				&route53.VPC{
					VPCId:     aws.String("vpc-8BAADF00D"),
					VPCRegion: aws.String("us-east-1"),
				},
			},
		},
		"zone2": route53.GetHostedZoneOutput{
			VPCs: []*route53.VPC{
				&route53.VPC{
					VPCId:     aws.String("vpc-1CEB00DA"),
					VPCRegion: aws.String("us-east-1"),
				},
			},
		},
		"zone3": route53.GetHostedZoneOutput{
			VPCs: []*route53.VPC{
				&route53.VPC{
					VPCId:     aws.String("vpc-C00010FF"),
					VPCRegion: aws.String("ap-south-1"),
				},
				&route53.VPC{
					VPCId:     aws.String("vpc-D15EA5E"),
					VPCRegion: aws.String("ap-south-1"),
				},
			},
		},
		"zone4": route53.GetHostedZoneOutput{
			VPCs: []*route53.VPC{
				&route53.VPC{
					VPCId:     aws.String("vpc-C00010FF"),
					VPCRegion: aws.String("sa-east-1"),
				},
			},
		},
		"zone5": route53.GetHostedZoneOutput{
			VPCs: []*route53.VPC{
				&route53.VPC{
					VPCId:     aws.String("vpc-DEADBAAD"),
					VPCRegion: aws.String("ap-south-1"),
				},
				&route53.VPC{
					VPCId:     aws.String("vpc-1CEB00DA"),
					VPCRegion: aws.String("ap-south-1"),
				},
			},
		},
	}

	return &mockRoute53Client{
		hostedZoneData: hostedZoneData,
	}
}

func setupNamespaceMocks(t *testing.T, publicNamespaces bool) serviceDiscoveryClient {
	// Mock Data
	namespaceType := servicediscovery.NamespaceTypeDnsPrivate
	if publicNamespaces {
		namespaceType = servicediscovery.NamespaceTypeDnsPublic
	}
	var namespaceData = map[string]servicediscovery.Namespace{
		"namespace1": servicediscovery.Namespace{
			Id:   aws.String("namespace1"),
			Name: aws.String("corp"),
			Properties: &servicediscovery.NamespaceProperties{
				DnsProperties: &servicediscovery.DnsProperties{
					HostedZoneId: aws.String("zone1"),
				},
			},
			Type: aws.String(namespaceType),
		},
		"namespace2": servicediscovery.Namespace{
			Id:   aws.String("namespace2"),
			Name: aws.String("prod"),
			Properties: &servicediscovery.NamespaceProperties{
				DnsProperties: &servicediscovery.DnsProperties{
					HostedZoneId: aws.String("zone2"),
				},
			},
			Type: aws.String(namespaceType),
		},
		"namespace3": servicediscovery.Namespace{
			Id:   aws.String("namespace3"),
			Name: aws.String("bridge"),
			Properties: &servicediscovery.NamespaceProperties{
				DnsProperties: &servicediscovery.DnsProperties{
					HostedZoneId: aws.String("zone3"),
				},
			},
			Type: aws.String(namespaceType),
		},
	}

	if !publicNamespaces {
		// Add in extra namespaces with the same name, since this is possible with private namespaces
		namespaceData["namespace4"] = servicediscovery.Namespace{
			Id:   aws.String("namespace4"),
			Name: aws.String("corp"),
			Properties: &servicediscovery.NamespaceProperties{
				DnsProperties: &servicediscovery.DnsProperties{
					HostedZoneId: aws.String("zone4"),
				},
			},
			Type: aws.String(namespaceType),
		}
		namespaceData["namespace5"] = servicediscovery.Namespace{
			Id:   aws.String("namespace5"),
			Name: aws.String("bridge"),
			Properties: &servicediscovery.NamespaceProperties{
				DnsProperties: &servicediscovery.DnsProperties{
					HostedZoneId: aws.String("zone5"),
				},
			},
			Type: aws.String(namespaceType),
		}

	}
	return &mockSDClient{
		usingFilter:      true,
		publicNamespaces: publicNamespaces,
		namespaceData:    namespaceData,
		t:                t,
	}

}
