// Copyright 2015-2017 Amazon.com, Inc. or its affiliates. All Rights Reserved.
//
// Licensed under the Apache License, Version 2.0 (the "License"). You may
// not use this file except in compliance with the License. A copy of the
// License is located at
//
//  http://aws.amazon.com/apache2.0/
//
// or in the "license" file accompanying this file. This file is distributed
// on an "AS IS" BASIS, WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either
// express or implied. See the License for the specific language governing
// permissions and limitations under the License.

package cloudformation

func GetSDSTemplate() string {
	return sds_template
}

var sds_template = `
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "AWS CloudFormation template to create a private DNS namespace to enable ECS Service Discovery.",
  "Parameters": {
    "SDSDescription": {
      "Type": "String",
      "Description": "Optional - The description of the private DNS namespace",
      "Default": "Created by the Amazon ECS CLI"
    },
    "SDSName": {
      "Type": "String",
      "Description": "The name of the Service Discovery Service",
      "Default": ""
    },
    "NamespaceID": {
      "Type": "String",
      "Description": "The ID of the namespace that you want to use for DNS configuration",
      "Default": ""
    },
    "DNSType": {
      "Type": "String",
      "Description": "The DNS type of the record that you want Route 53 to create.",
      "Default": ""
    },
    "DNSTTL": {
      "Type": "String",
      "Description": "The amount of time, in seconds, that you want DNS resolvers to cache the settings for this record.",
      "Default": "60"
    },
    "FailureThreshold": {
      "Type": "Number",
      "Description": "The number of 30-second intervals that you want service discovery to wait after receiving an UpdateInstanceCustomHealthStatus request before it changes the health status of a service instance.",
      "Default": 1
    }
  },
  "Resources": {
    "ServiceDiscoveryService": {
      "Type" : "AWS::ServiceDiscovery::Service",
      "Properties" : {
        "Description" : { "Ref" : "SDSDescription" },
        "DnsConfig" : {
          "DnsRecords" : [ {
            "Type" : { "Ref" : "DNSType" },
            "TTL" : { "Ref" : "DNSTTL" }
          } ],
          "NamespaceId" : { "Ref" : "NamespaceID" }
        },
        "HealthCheckCustomConfig" : {
          "FailureThreshold" : { "Ref" : "FailureThreshold" }
        },
        "Name" : { "Ref" : "SDSName" }
      }
    }
  },
  "Outputs" : {
    "ServiceDiscoveryServiceARN" : {
      "Description": "The ARN of the Service Discovery Service which can be used when launching an ECS Service.",
      "Value" : { "Fn::GetAtt" : [ "ServiceDiscoveryService", "Arn" ]}
    }
  }
}
`
