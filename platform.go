package cloudformation

type Platform interface {
	CreateResource(resourcetype string, input []byte)
}
