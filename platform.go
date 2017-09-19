package cloudformation

import (
	"errors"
	"github.com/the-no/aws-sdk-go/aws/session"
	"github.com/the-no/aws-sdk-go/service/ec2"
)

type Platform interface {
	//CreateResource(resourcetype string, input []byte)
	NewClinet(typ string, s *session.Session) (ClinetInterface, error)
	PseudoParameter(name string, req *Request) (interface{}, error)
}

type AWSPlatform struct {
}

func (c *AWSPlatform) NewClinet(typ string, s *session.Session) ClinetInterface {
	switch typ {
	case "EC2":
		return ec2.New(s)
	}
	return nil
}

func (c *AWSPlatform) PseudoParameter(name string, req *Request) (interface{}, error) {

	switch name {
	case "AWS::Region":
		return "us-west-2", nil
	case "AWS::StackId":
	case "AWS::StackName":
		return req.StackName, nil
	case "AWS::NoValue":
		return nil, nil
		//	case "AWS::NotificationARNs":
	case "AWS::AccountId":
	}
	return "", errors.New("Invail Parameter")
}
