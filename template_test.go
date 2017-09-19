package cloudformation

import (
	"encoding/json"
	//"fmt"
	. "gopkg.in/check.v1"
	"testing"
	"time"
)

var inputbuf = `{
  "AWSTemplateFormatVersion" : "2010-09-09",

  "Mappings" : {
    "RegionMap" : {
      "us-east-1"      : { "AMI" : "ami-7f418316", "TestAz" : "us-east-1a" },
      "us-west-1"      : { "AMI" : "ami-951945d0", "TestAz" : "us-west-1a" },
      "us-west-2"      : { "AMI" : "ami-16fd7026", "TestAz" : "us-west-2a" },
      "eu-west-1"      : { "AMI" : "ami-24506250", "TestAz" : "eu-west-1a" },
      "sa-east-1"      : { "AMI" : "ami-3e3be423", "TestAz" : "sa-east-1a" },
      "ap-southeast-1" : { "AMI" : "ami-74dda626", "TestAz" : "ap-southeast-1a" },
      "ap-southeast-2" : { "AMI" : "ami-b3990e89", "TestAz" : "ap-southeast-2a" },
      "ap-northeast-1" : { "AMI" : "ami-dcfa4edd", "TestAz" : "ap-northeast-1a" }
    }
  },
    
  "Parameters" : {
    "EnvType" : {
      "Description" : "Environment type.",
      "Default" : "test",
      "Type" : "String",
      "AllowedValues" : ["prod", "test"],
      "ConstraintDescription" : "must specify prod or test."
    }
  },
  
  "Conditions" : {
    "CreateProdResources" : {"Fn::Equals" : [{"Ref" : "EnvType"}, "prod"]}
  },
  
  "Resources" : {
    "EC2Instance" : {
      "Type" : "AWS::EC2::Instance",
      "Properties" : {
        "ImageId" : { "Fn::FindInMap" : [ "RegionMap", { "Ref" : "AWS::Region" }, "AMI" ]}
      }
    },
    
    "MountPoint" : {
      "Type" : "AWS::EC2::VolumeAttachment",
      "Condition" : "CreateProdResources",
      "Properties" : {
        "InstanceId" : { "Ref" : "EC2Instance" },
        "VolumeId"  : { "Ref" : "NewVolume" },
        "Device" : "/dev/sdh"
      }
    },

    "NewVolume" : {
      "Type" : "AWS::EC2::Volume",
      "Condition" : "CreateProdResources",
      "Properties" : {
        "Size" : "100",
        "AvailabilityZone" : { "Fn::GetAtt" : [ "EC2Instance", "AvailabilityZone" ]}
      }
    }
  },
  
  "Outputs" : {
    "VolumeId" : {
      "Value" : { "Ref" : "NewVolume" }, 
      "Condition" : "CreateProdResources"
    }
  }  
}`

var requestbuf = `{
    "capabilities": [],
    "disableRollback": false,
    "notificationARNs": [],
    "parameters": [{
        "parameterKey": "EnvType",
        "parameterValue": "prod"
    }],
    "roleARN": null,
    "stackName": "test",
    "tags": [{
        "key": "shu",
        "value": "mh"
    }],
    "templateURL": "https://s3-us-west-2.amazonaws.com/cloudformation-templates-us-west-2/WordPress_Single_Instance.template",
    "timeoutInMinutes": null,
    "stackPolicyBody": null,
    "stackPolicyURL": null,
    "clientRequestToken": "Console-CreateStack-da7c5148-7de0-40fe-bc40-ac2f4e163982"
}
`

type TemplateTest struct{}

var _ = Suite(&TemplateTest{})

func (testSuite *TemplateTest) TestTemplate(c *C) {
	tmp := NewTemplate()
	err := json.Unmarshal([]byte(inputbuf), tmp)
	c.Assert(err, IsNil)
	for k, v := range tmp.Conditions {
		c.Log(k, v)
	}
}
func TestTemplate1(t *testing.T) {
	tmp := NewTemplate()
	err := json.Unmarshal([]byte(inputbuf), tmp)
	t.Log("---------", err)
	/*	for k, v := range tmp.Resources {
		t.Log(k, v.Type, string(v.Properties))
	}*/

	/*	for k, v := range tmp.Conditions {
		t.Logf("%s ,%#v", k, v)
	}*/

	req := &Request{}
	err = json.Unmarshal([]byte(requestbuf), req)
	t.Log("---------", err)
	fm, err := tmp.CreateFormation(req)
	t.Logf("%#v,\n%#v\n", fm.Conditions, fm.Parameters)

	fm.StartResourceUnits()
	//data, err := json.Marshal(tmp)
	//	t.Log(string(data))
	time.Sleep(20 * time.Second)
}
