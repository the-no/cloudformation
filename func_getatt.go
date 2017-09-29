package cloudformation

import "encoding/json"

// GetAtt returns a new instance of GetAttFunc.
func GetAtt(resource, name string) *StringExpr {
	return GetAttFunc{Resource: resource, Name: name}.String()
}

// GetAttFunc represents an invocation of the Fn::GetAtt intrinsic.
//
// The intrinsic function Fn::GetAtt returns the value of an attribute from a
// resource in the template.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-getatt.html
type GetAttFunc struct {
	Resource string
	Name     StringExpr
}

// MarshalJSON returns a JSON representation of the object
func (f GetAttFunc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		FnGetAtt []string `json:"Fn::GetAtt"`
	}{FnGetAtt: [][]interface{}{f.Resource, f.Name}})
}

// UnmarshalJSON sets the object from the provided JSON representation
func (f *GetAttFunc) UnmarshalJSON(data []byte) error {
	v := struct {
		FnGetAtt [2]json.RawMessage `json:"Fn::GetAtt"`
	}{}
	if err := json.Unmarshal(data, &v); err != nil {
		return err
	}
	if len(v.FnGetAtt) != 2 {
		return &json.UnsupportedValueError{Str: string(data)}
	}

	if err := json.Unmarshal(v.FnGetAtt[0], &f.Resource); err != nil {
		return err
	}
	if err := json.Unmarshal(v.FnGetAtt[1], &f.Name); err != nil {
		return err
	}
	return nil
}

func (f GetAttFunc) String() *StringExpr {
	return &StringExpr{Func: f}
}

func (f GetAttFunc) Exec(fm *Formation) (interface{}, error) {

	if v, ok := fm.Resources[f.Resource]; ok {
		name := f.Name.Literal
		if f.Name.Func != nil {
			v, err := f.Name.Func.Exec(fm)
			if err != nil {
				return nil, err
			}
			name = v.(string)
		}
		if err := v.Wait(); err != nil {
			return v.Attr(name), err
		}
		return r, nil
	}
	return nil, nil
}

/*func (r GetAttFunc) DependResource() []string {
	return nil
}*/

var _ StringFunc = GetAttFunc{} // GetAttFunc must implement StringFunc
