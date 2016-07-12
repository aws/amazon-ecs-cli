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

package cloudformation

func GetTemplate() string {
	return template
}

// TODO: Improvements:
// 1. Auto detect default vpc
// 2. Auto detect existing key pairs
// 3. Create key pair when none exist
// 4. Remove the hardcoded 2 subnets creation
var template = `
{
  "AWSTemplateFormatVersion": "2010-09-09",
  "Description": "AWS CloudFormation template to create resources required to run tasks on an ECS cluster.",
  "Mappings": {
    "VpcCidrs": {
      "vpc": {"cidr" : "10.0.0.0/16"},
      "pubsubnet1": {"cidr" : "10.0.0.0/24"},
      "pubsubnet2": {"cidr" :"10.0.1.0/24"}
    }
  },
  "Parameters": {
    "EcsAmiId": {
      "Type": "String",
      "Description": "ECS EC2 AMI id",
      "Default": ""
    },
    "EcsInstanceType": {
      "Type": "String",
      "Description": "ECS EC2 instance type",
      "Default": "t2.micro",
      "AllowedValues": [
        "t2.nano",
        "t2.micro",
        "t2.small",
        "t2.medium",
        "t2.large",
        "m3.medium",
        "m3.large",
        "m3.xlarge",
        "m3.2xlarge",
        "m4.large",
        "m4.xlarge",
        "m4.2xlarge",
        "m4.4xlarge",
        "m4.10xlarge",
        "c4.large",
        "c4.xlarge",
        "c4.2xlarge",
        "c4.4xlarge",
        "c4.8xlarge",
        "c3.large",
        "c3.xlarge",
        "c3.2xlarge",
        "c3.4xlarge",
        "c3.8xlarge",
        "r3.large",
        "r3.xlarge",
        "r3.2xlarge",
        "r3.4xlarge",
        "r3.8xlarge",
        "i2.xlarge",
        "i2.2xlarge",
        "i2.4xlarge",
        "i2.8xlarge",
        "g2.2xlarge",
        "g2.8xlarge",
        "d2.xlarge",
        "d2.2xlarge",
        "d2.4xlarge",
        "d2.8xlarge"
      ],
      "ConstraintDescription": "must be a valid EC2 instance type."
    },
    "KeyName": {
      "Type": "AWS::EC2::KeyPair::KeyName",
      "Description": "Optional - Name of an existing EC2 KeyPair to enable SSH access to the ECS instances",
      "Default": ""
    },
    "VpcId": {
      "Type": "String",
      "Description": "Optional - VPC Id of existing VPC. Leave blank to have a new VPC created",
      "Default": "",
      "AllowedPattern": "^(?:vpc-[0-9a-f]{8}|)$",
      "ConstraintDescription": "VPC Id must begin with 'vpc-' or leave blank to have a new VPC created"
    },
    "SubnetIds": {
      "Type": "CommaDelimitedList",
      "Description": "Optional - Comma separated list of two (2) existing VPC Subnet Ids where ECS instances will run.  Required if setting VpcId.",
      "Default": ""
    },
    "AsgMaxSize": {
      "Type": "Number",
      "Description": "Maximum size and initial Desired Capacity of ECS Auto Scaling Group",
      "Default": "1"
    },
    "SecurityGroup": {
      "Type": "String",
      "Description": "Optional - Existing security group to associate the container instances. Creates one by default.",
      "Default": ""
    },
    "SourceCidr": {
      "Type": "String",
      "Description": "Optional - CIDR/IP range for EcsPort - defaults to 0.0.0.0/0",
      "Default": "0.0.0.0/0"
    },
    "EcsPort" : {
      "Type" : "String",
      "Description" : "Optional - Security Group port to open on ECS instances - defaults to port 80",
      "Default" : "80"
    },
    "VpcAvailabilityZones": {
      "Type": "CommaDelimitedList",
      "Description": "Optional - Comma-delimited list of VPC availability zones in which to create subnets.  Required if setting VpcId.",
      "Default": ""
    },
    "EcsCluster" : {
      "Type" : "String",
      "Description" : "ECS Cluster Name",
      "Default" : "default"
    }
  },
  "Conditions": {
    "CreateVpcResources": {
      "Fn::Equals": [
        {
          "Ref": "VpcId"
        },
        ""
      ]
    },
    "CreateSecurityGroup": {
      "Fn::Equals": [
        {
          "Ref": "SecurityGroup"
        },
        ""
      ]
    },
    "CreateEC2LCWithKeyPair": {
      "Fn::Not": [
        {
          "Fn::Equals": [
            {
              "Ref": "KeyName"
            },
            ""
          ]
        }
      ]
    },
    "CreateEC2LCWithoutKeyPair": {
      "Fn::Equals": [
        {
          "Ref": "KeyName"
        },
        ""
      ]
    },
    "UseSpecifiedVpcAvailabilityZones": {
      "Fn::Not": [
        {
          "Fn::Equals": [
            {
              "Fn::Join": [
                "",
                {
                  "Ref": "VpcAvailabilityZones"
                }
              ]
            },
            ""
          ]
        }
      ]
    }
  },
  "Resources": {
    "Vpc": {
      "Condition": "CreateVpcResources",
      "Type": "AWS::EC2::VPC",
      "Properties": {
        "CidrBlock": {
          "Fn::FindInMap": ["VpcCidrs", "vpc", "cidr"]
        }
      }
    },
    "PubSubnetAz1": {
      "Condition": "CreateVpcResources",
      "Type": "AWS::EC2::Subnet",
      "Properties": {
        "VpcId": {
          "Ref": "Vpc"
        },
        "CidrBlock": {
          "Fn::FindInMap": ["VpcCidrs", "pubsubnet1", "cidr"]
        },
        "AvailabilityZone": {
          "Fn::If": [
            "UseSpecifiedVpcAvailabilityZones",
            {
              "Fn::Select": [
                "0",
                {
                  "Ref": "VpcAvailabilityZones"
                }
              ]
            },
            {
              "Fn::Select": [
                "0",
                {
                  "Fn::GetAZs": {
                    "Ref": "AWS::Region"
                  }
                }
              ]
            }
          ]
        }
      }
    },
    "PubSubnetAz2": {
      "Condition": "CreateVpcResources",
      "Type": "AWS::EC2::Subnet",
      "Properties": {
        "VpcId": {
          "Ref": "Vpc"
        },
        "CidrBlock": {
          "Fn::FindInMap": ["VpcCidrs", "pubsubnet2", "cidr"]
        },
        "AvailabilityZone": {
          "Fn::If": [
            "UseSpecifiedVpcAvailabilityZones",
            {
              "Fn::Select": [
                "1",
                {
                  "Ref": "VpcAvailabilityZones"
                }
              ]
            },
            {
              "Fn::Select": [
                "1",
                {
                  "Fn::GetAZs": {
                    "Ref": "AWS::Region"
                  }
                }
              ]
            }
          ]
        }
      }
    },
    "InternetGateway": {
      "Condition": "CreateVpcResources",
      "Type": "AWS::EC2::InternetGateway"
    },
    "AttachGateway": {
      "Condition": "CreateVpcResources",
      "Type": "AWS::EC2::VPCGatewayAttachment",
      "Properties": {
        "VpcId": {
          "Ref": "Vpc"
        },
        "InternetGatewayId": {
          "Ref": "InternetGateway"
        }
      }
    },
    "RouteViaIgw": {
      "Condition": "CreateVpcResources",
      "Type": "AWS::EC2::RouteTable",
      "Properties": {
        "VpcId": {
          "Ref": "Vpc"
        }
      }
    },
    "PublicRouteViaIgw": {
      "Condition": "CreateVpcResources",
      "DependsOn": "AttachGateway",
      "Type": "AWS::EC2::Route",
      "Properties": {
        "RouteTableId": {
          "Ref": "RouteViaIgw"
        },
        "DestinationCidrBlock": "0.0.0.0/0",
        "GatewayId": {
          "Ref": "InternetGateway"
        }
      }
    },
    "PubSubnet1RouteTableAssociation": {
      "Condition": "CreateVpcResources",
      "Type": "AWS::EC2::SubnetRouteTableAssociation",
      "Properties": {
        "SubnetId": {
          "Ref": "PubSubnetAz1"
        },
        "RouteTableId": {
          "Ref": "RouteViaIgw"
        }
      }
    },
    "PubSubnet2RouteTableAssociation": {
      "Condition": "CreateVpcResources",
      "Type": "AWS::EC2::SubnetRouteTableAssociation",
      "Properties": {
        "SubnetId": {
          "Ref": "PubSubnetAz2"
        },
        "RouteTableId": {
          "Ref": "RouteViaIgw"
        }
      }
    },
    "EcsSecurityGroup": {
      "Condition": "CreateSecurityGroup",
      "Type": "AWS::EC2::SecurityGroup",
      "Properties": {
        "GroupDescription": "ECS Allowed Ports",
        "VpcId": {
          "Fn::If": [
            "CreateVpcResources",
            {
              "Ref": "Vpc"
            },
            {
              "Ref": "VpcId"
            }
          ]
        },
        "SecurityGroupIngress" : [ {
            "IpProtocol" : "tcp",
            "FromPort" : { "Ref" : "EcsPort" },
            "ToPort" : { "Ref" : "EcsPort" },
            "CidrIp" : { "Ref" : "SourceCidr" } 
        } ]
      }
    },
    "EcsInstancePolicy": {
      "Type": "AWS::IAM::Role",
      "Properties": {
        "AssumeRolePolicyDocument": {
          "Version": "2012-10-17",
          "Statement": [
            {
              "Effect": "Allow",
              "Principal": {
                "Service": [
                  "ec2.amazonaws.com"
                ]
              },
              "Action": [
                "sts:AssumeRole"
              ]
            }
          ]
        },
        "Path": "/",
        "ManagedPolicyArns": [
          "arn:aws:iam::aws:policy/service-role/AmazonEC2ContainerServiceforEC2Role"
        ]
      }
    },
    "EcsInstanceProfile": {
      "Type": "AWS::IAM::InstanceProfile",
      "Properties": {
        "Path": "/",
        "Roles": [
          {
            "Ref": "EcsInstancePolicy"
          }
        ]
      }
    },
    "EcsInstanceLc": {
      "Condition": "CreateEC2LCWithKeyPair",
      "Type": "AWS::AutoScaling::LaunchConfiguration",
      "Properties": {
	"ImageId": { "Ref" : "EcsAmiId" },
        "InstanceType": {
          "Ref": "EcsInstanceType"
        },
        "AssociatePublicIpAddress": true,
        "IamInstanceProfile": {
          "Ref": "EcsInstanceProfile"
        },
        "KeyName": {
          "Ref": "KeyName"
        },
        "SecurityGroups": {
          "Fn::If": [
            "CreateSecurityGroup",
            [ {
              "Ref": "EcsSecurityGroup"
            } ],
            [ {
              "Ref": "SecurityGroup"
            } ]
          ]
        },
        "UserData": {
          "Fn::Base64": {
            "Fn::Join": [
              "",
              [
                "#!/bin/bash\n",
                "echo ECS_CLUSTER=",
                {
                  "Ref": "EcsCluster"
                },
                " >> /etc/ecs/ecs.config\n"
              ]
            ]
          }
        }
      }
    },
    "EcsInstanceLcWithoutKeyPair": {
      "Condition": "CreateEC2LCWithoutKeyPair",
      "Type": "AWS::AutoScaling::LaunchConfiguration",
      "Properties": {
	"ImageId": { "Ref" : "EcsAmiId" },
        "InstanceType": {
          "Ref": "EcsInstanceType"
        },
        "AssociatePublicIpAddress": true,
        "IamInstanceProfile": {
          "Ref": "EcsInstanceProfile"
        },
        "SecurityGroups": {
          "Fn::If": [
            "CreateSecurityGroup",
            [ {
              "Ref": "EcsSecurityGroup"
            } ],
            [ {
              "Ref": "SecurityGroup"
            } ]
          ]
        },
        "UserData": {
          "Fn::Base64": {
            "Fn::Join": [
              "",
              [
                "#!/bin/bash\n",
                "echo ECS_CLUSTER=",
                {
                  "Ref": "EcsCluster"
                },
                " >> /etc/ecs/ecs.config\n"
              ]
            ]
          }
        }
      }
    },
    "EcsInstanceAsg": {
      "Type": "AWS::AutoScaling::AutoScalingGroup",
      "Properties": {
        "VPCZoneIdentifier": {
          "Fn::If": [
            "CreateVpcResources",
            [
              {
                "Fn::Join": [
                  ",",
                  [
                    {
                      "Ref": "PubSubnetAz1"
                    },
                    {
                      "Ref": "PubSubnetAz2"
                    }
                  ]
                ]
              }
            ],
            {
              "Ref": "SubnetIds"
            }
          ]
        },
        "LaunchConfigurationName": {
          "Fn::If": [
            "CreateEC2LCWithKeyPair",
            {
              "Ref": "EcsInstanceLc"
            },
            {
              "Ref": "EcsInstanceLcWithoutKeyPair"
            }
          ]
        },
        "MinSize": "1",
        "MaxSize": {
          "Ref": "AsgMaxSize"
        },
        "DesiredCapacity": {
          "Ref": "AsgMaxSize"
        },
        "Tags": [
          {
            "Key": "Name",
            "Value": {
              "Fn::Join": [
                "",
                [
                  "ECS Instance - ",
                  {
                    "Ref": "AWS::StackName"
                  }
                ]
              ]
            },
            "PropagateAtLaunch": "true"
          }
        ]
      }
    }
  }
}
`
