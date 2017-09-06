package cloudformation

import "encoding/json"

//import "reflect"

// If returns a new instance of IfFunc for the provided string expressions.
//
// See also: IfList
func Equalse() EqualsFunc {
	return EqualsFunc{
		Value1: &StringExpr{},
		Value2: &StringExpr{},
	}
}

// IfList returns a new instance of IfFunc for the provided string list expressions.
//
// See also: If
/*func IfList(condition string, valueIfTrue, valueIfFalse StringListable) IfFunc {
	return IfFunc{
		list:         true,
		Condition:    condition,
		ValueIfTrue:  *valueIfTrue.StringList(),
		ValueIfFalse: *valueIfFalse.StringList(),
	}
}*/

// IfFunc represents an invocation of the Fn::If intrinsic.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-conditions.html
type EqualsFunc struct {
	Value1 *StringExpr
	Value2 *StringExpr
}

// MarshalJSON returns a JSON representation of the object
func (f EqualsFunc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		FnEquals []*StringExpr `json:"Fn::Equals"`
	}{FnEquals: []*StringExpr{f.Value1, f.Value2}})
}

// UnmarshalJSON sets the object from the provided JSON representation
func (f *EqualsFunc) UnmarshalJSON(buf []byte) error {
	v := struct {
		FnEquals []*StringExpr `json:"Fn::Equals"`
	}{}
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}

	if len(v.FnEquals) != 2 {
		return &json.UnsupportedValueError{Str: string(buf)}
	}
	f.Value1 = v.FnEquals[0]
	f.Value2 = v.FnEquals[1]

	return nil
}

func (f EqualsFunc) String() *StringExpr {
	return &StringExpr{Func: f}
}

func (f EqualsFunc) Bool() *BoolExpr {
	return &BoolExpr{Func: f}
}

func (f EqualsFunc) Exec(fm *Formation) (interface{}, error) {

	v1, err := f.Value1.evalExpr(fm)
	if err != nil {
		return false, err
	}

	v2, err := f.Value2.evalExpr(fm)
	if err != nil {
		return false, err
	}
	return v1 == v2, nil
}

func (r EqualsFunc) DependResource(fm *Formation) []*ResourceUnit {
	dps := []*ResourceUnit{}
	if r.Value1.Func != nil {
		dps = append(dps, r.Value1.Func.DependResource(fm)...)
	}
	if r.Value2.Func != nil {
		dps = append(dps, r.Value1.Func.DependResource(fm)...)
	}
	return dps
}

var _ StringFunc = EqualsFunc{} // EqualsFunc must implement StringFunc
var _ BoolFunc = EqualsFunc{}   // EqualsFunc must implement BoolFunc
