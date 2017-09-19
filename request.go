package cloudformation

type Request struct {
	Capabilities     []string `json:"capabilities"`
	Platform         string   `json:"platform"`
	DisableRollback  bool     `json:"disableRollback"`
	NotificationARNs []string `json:"notificationARNs"`
	Parameters       []struct {
		ParameterKey   string `json:"parameterKey"`
		ParameterValue string `json:"parameterValue"`
	} `json:"parameters"`
	RoleARN   string `json:"roleARN"`
	StackName string `json:"stackName"`
	Tags      []struct {
		Key   string `json:"key"`
		Value string `json:"value"`
	} `json:"tags"`
	TemplateURL        string `json:"templateURL"` // "https://s3-us-west-2.amazonaws.com/cloudformation-templates-us-west-2/WordPress_Single_Instance.template",
	TimeoutInMinutes   string `json:"timeoutInMinutes"`
	StackPolicyBody    string `json:"stackPolicyBody"`
	StackPolicyURL     string `json:"stackPolicyURL"`
	ClientRequestToken string `json:"clientRequestToken"` // "Console-CreateStack-da7c5148-7de0-40fe-bc40-ac2f4e163982"
}
