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

func GetPrivateNamespaceTemplate() string {
	return private_namespace_template
}

var private_namespace_template = `
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "AWS CloudFormation template to create a private DNS namespace to enable ECS Service Discovery.",
  "Parameters": {
    "NamespaceDescription": {
      "Type": "String",
      "Description": "Optional - The description of the private DNS namespace",
      "Default": "Created by the Amazon ECS CLI"
    },
    "VPCID": {
      "Type": "String",
      "Description": "The VPC attached to the DNS private namespace",
      "Default": ""
    },
    "NamespaceName": {
      "Type": "String",
      "Description": "The name of the namespace",
      "Default": ""
    }
  },
  "Resources": {
    "PrivateDNSNamespace": {
      "Type" : "AWS::ServiceDiscovery::PrivateDnsNamespace",
      "Properties" : {
        "Description" : { "Ref" : "NamespaceDescription" },
        "Vpc" : { "Ref" : "VPCID" },
        "Name" : { "Ref" : "NamespaceName" }
      }
    }
  },
  "Outputs" : {
    "PrivateDNSNamespaceID" : {
      "Description": "The ID of the private DNS namespace.",
      "Value" : { "Fn::GetAtt" : [ "PrivateDNSNamespace", "Id" ]}
    }
  }
}
`
