package cloudformation

import (
	"github.com/the-no/aws-sdk-go/aws"
)

type ClinetInterface interface {
	CreateResource(typ string, data []byte) (intput, output interface{}, ref aws.Referencer, err error)
}
