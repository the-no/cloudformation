package cloudformation

import (
	"encoding/json"
	"errors"
	"github.com/the-no/aws-sdk-go/aws/session"
	"regexp"
)

// NewTemplate returns a new empty Template initialized with some
// default values.
func NewTemplate() *Template {
	return &Template{
		AWSTemplateFormatVersion: "2010-09-09",
		Mappings:                 map[string]*Mapping{},
		Parameters:               map[string]*ParameterChecker{},
		Resources:                map[string]*Resource{},
		Outputs:                  map[string]*Output{},
		Conditions:               map[string]*BoolExpr{},
	}
}

// Template represents a cloudformation template.
type Template struct {
	AWSTemplateFormatVersion string                       `json:",omitempty"`
	Description              string                       `json:",omitempty"`
	Mappings                 map[string]*Mapping          `json:",omitempty"`
	Parameters               map[string]*ParameterChecker `json:",omitempty"`
	Resources                map[string]*Resource         `json:",omitempty"`
	Outputs                  map[string]*Output           `json:",omitempty"`
	Conditions               map[string]*BoolExpr         `json:",omitempty"`
}

func (t *Template) CreateFormation(req *Request, s *session.Session) (*Formation, error) {
	fm := &Formation{
		Parameters: make(map[string]*ParameterValue),
		Resources:  make(map[string]*ResourceUnit),
		Conditions: make(map[string]bool),
		Mappings:   t.Mappings,
		Platform:   getplatform(req.Platform),
		Session:    s,
	}

	for _, v := range req.Parameters {
		input := &ParameterValue{
			Name:  v.ParameterKey,
			Value: v.ParameterValue,
		}
		if p, ok := t.Parameters[v.ParameterKey]; ok {
			input.Type = p.Type
			fm.Parameters[v.ParameterKey] = input
		}
	}

	if err := t.evalConditions(fm); err != nil {
		return nil, err
	}

	for k, r := range t.Resources {
		for _, d := range r.DependsOn {
			if _, ok := t.Resources[d]; !ok {
				return nil, errors.New("Found Resource [" + d + "] Faild!")
			}
		}
		fm.Resources[k] = NewResourceUnit(fm, k, r)
	}
	return fm, nil
}

func (t *Template) chechParamer(fm *Formation) error {
	for k, v := range t.Parameters {
		if _, ok := fm.Parameters[k]; ok {
			v.Check(fm, k)
		} else {
			fm.Parameters[k] = &ParameterValue{
				Type:  v.Type,
				Name:  k,
				Value: v.Default,
			}
		}
	}
	return nil
}

func (t *Template) evalConditions(fm *Formation) error {
	for k, v := range t.Conditions {
		result, err := v.evalExpr(fm)
		if err != nil {
			return err
		}
		fm.Conditions[k] = result
	}
	return nil
}

// AddResource adds the resource to the template as name, displacing
// any resource with the same name that already exists.
/*func (t *Template) AddResource(name string, resource ResourceProperties) *Resource {
	templateResource := &Resource{Properties: resource}
	t.Resources[name] = templateResource
	return templateResource
}*/

// Mapping matches a key to a corresponding set of named values. For example,
// if you want to set values based on a region, you can create a mapping that
// uses the region name as a key and contains the values you want to specify
// for each specific region. You use the Fn::FindInMap intrinsic function to
// retrieve values in a map.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/mappings-section-structure.html
type Mapping map[string]map[string]string

/*type Mapping struct {
	l sync.RWMutex

}*/

/*type Mapping struct {
	lock sync.RWMutex
	map[string]map[string]string
}
*/
// Parameter represents a parameter to the template.
//
// You can use the optional Parameters section to pass values into your
// template when you create a stack. With parameters, you can create templates
// that are customized each time you create a stack. Each parameter must
// contain a value when you create a stack. You can specify a default value to
// make the parameter optional.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/parameters-section-structure.html
type ParameterChecker struct {
	Type                  string        `json:",omitempty"`
	Default               string        `json:",omitempty"`
	NoEcho                *BoolExpr     `json:",omitempty"`
	AllowedValues         []interface{} `json:",omitempty"`
	AllowedPattern        string        `json:",omitempty"`
	MinLength             *IntegerExpr  `json:",omitempty"`
	MaxLength             *IntegerExpr  `json:",omitempty"`
	MinValue              *IntegerExpr  `json:",omitempty"`
	MaxValue              *IntegerExpr  `json:",omitempty"`
	Description           string        `json:",omitempty"`
	ConstraintDescription string        `json:",omitempty"`
}

func (self *ParameterChecker) Check(fm *Formation, param string) bool {
	switch self.Type {
	case "String":
		return checkString(fm, param, self)
	case "Number":
		return checkNumber(fm, param, self)
	default:
	}
	return true
}

func checkString(fm *Formation, param string, ch *ParameterChecker) bool {

	paramtervalue := fm.Parameters[param].Value
	if len(ch.AllowedPattern) > 0 {
		return regexp.MustCompile(ch.AllowedPattern).MatchString(paramtervalue)
	}
	for _, v := range ch.AllowedValues {
		if v == paramtervalue {
			return true
		}
	}

	if ch.MaxLength != nil {
		maxlength, err := ch.MaxLength.evalExpr(fm)
		if err != nil {
			return false
		}
		if len(paramtervalue) > int(maxlength) {
			return false
		}
	}

	if ch.MinLength != nil {
		minlength, err := ch.MinLength.evalExpr(fm)
		if err != nil {
			return false
		}
		if len(paramtervalue) < int(minlength) {
			return false
		}
	}

	return false
}

func checkNumber(fm *Formation, param string, ch *ParameterChecker) bool {
	/*paramtervalue := fm.Parameters[param].Value
	for _, v := range ch.AllowedValues {
		val := v.(int64)
		if paramtervalue == val {
			return true
		}
	}

	if ch.MaxLength != nil {
		maxlength, err := fm.evalIntegerExpr(ch.MaxLength)
		if err != nil {
			return false
		}
		if maxlength > 0 && paramtervalue > maxlength {
			return false
		}
	}

	if ch.MinLength != nil {
		minlength, err := fm.evalIntegerExpr(ch.MinLength)
		if err != nil {
			return false
		}
		if paramtervalue < minlength {
			return false
		}
	}*/

	return true
}

func checkResourceType(fm *Formation, param string, ch *ParameterChecker) bool {
	return true
}

// OutputExport represents the name of the resource output that should
// be used for cross stack references.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/walkthrough-crossstackref.html
type OutputExport struct {
	Name Stringable `json:",omitempty"`
}

// Output represents a template output
//
// The optional Outputs section declares output values that you want to view from the
// AWS CloudFormation console or that you want to return in response to describe stack calls.
// For example, you can output the Amazon S3 bucket name for a stack so that you can easily find it.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/outputs-section-structure.html
type Output struct {
	Description string        `json:",omitempty"`
	Value       interface{}   `json:",omitempty"`
	Export      *OutputExport `json:",omitempty"`
}

// ResourceProperties is an interface that is implemented by resource objects.
type ResourceProperties interface {
	CfnResourceType() string
}

// Resource represents a resource in a cloudformation template. It contains resource
// metadata and, in Properties, a struct that implements ResourceProperties which
// contains the properties of the resource.
type Resource struct {
	CreationPolicy *CreationPolicy        `json:",omitempty"`
	DeletionPolicy string                 `json:",omitempty"`
	DependsOn      []string               `json:",omitempty"`
	Metadata       map[string]interface{} `json:",omitempty"`
	UpdatePolicy   *UpdatePolicy          `json:",omitempty"`
	Condition      string                 `json:",omitempty"`
	Type           string                 `json:",omitempty"`
	Properties     json.RawMessage        `json:",omitempty"`
}

// MarshalJSON returns a JSON representation of the object
/*func (r Resource) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		Type           string
		CreationPolicy *CreationPolicy        `json:",omitempty"`
		DeletionPolicy string                 `json:",omitempty"`
		DependsOn      []string               `json:",omitempty"`
		Metadata       map[string]interface{} `json:",omitempty"`
		UpdatePolicy   *UpdatePolicy          `json:",omitempty"`
		Condition      string                 `json:",omitempty"`
		Properties     ResourceProperties
	}{
		Type:           r.Properties.CfnResourceType(),
		CreationPolicy: r.CreationPolicy,
		DeletionPolicy: r.DeletionPolicy,
		DependsOn:      r.DependsOn,
		Metadata:       r.Metadata,
		UpdatePolicy:   r.UpdatePolicy,
		Condition:      r.Condition,
		Properties:     r.Properties,
	})
}

// UnmarshalJSON sets the object from the provided JSON representation
func (r *Resource) UnmarshalJSON(buf []byte) error {
	m := map[string]interface{}{}
	if err := json.Unmarshal(buf, &m); err != nil {
		return err
	}

	typeName := m["Type"].(string)
	r.DependsOn, _ = m["DependsOn"].([]string)
	r.Metadata, _ = m["Metadata"].(map[string]interface{})
	r.DeletionPolicy, _ = m["DeletionPolicy"].(string)
	r.Properties = NewResourceByType(typeName)
	if r.Properties == nil {
		return fmt.Errorf("unknown resource type: %s", typeName)
	}

	propertiesBuf, err := json.Marshal(m["Properties"])
	if err != nil {
		return err
	}
	if err := json.Unmarshal(propertiesBuf, r.Properties); err != nil {
		return err
	}
	return nil
}*/
