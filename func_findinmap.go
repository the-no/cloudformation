package cloudformation

import (
	"encoding/json"
	"errors"
	//"fmt"
)

// FindInMap returns a new instance of FindInMapFunc.
func FindInMap(mapName string, topLevelKey Stringable, secondLevelKey Stringable) *StringExpr {
	return FindInMapFunc{
		MapName:        mapName,
		TopLevelKey:    *topLevelKey.String(),
		SecondLevelKey: *secondLevelKey.String(),
	}.String()
}

// FindInMapFunc represents an invocation of the Fn::FindInMap intrinsic.
//
// The intrinsic function Fn::FindInMap returns the value corresponding to
// keys in a two-level map that is declared in the Mappings section.
//
// See http://docs.aws.amazon.com/AWSCloudFormation/latest/UserGuide/intrinsic-function-reference-findinmap.html
type FindInMapFunc struct {
	MapName        string
	TopLevelKey    StringExpr
	SecondLevelKey StringExpr
}

// MarshalJSON returns a JSON representation of the object
func (f FindInMapFunc) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		FnFindInMap []interface{} `json:"Fn::FindInMap"`
	}{FnFindInMap: []interface{}{f.MapName, f.TopLevelKey, f.SecondLevelKey}})
}

// UnmarshalJSON sets the object from the provided JSON representation
func (f *FindInMapFunc) UnmarshalJSON(buf []byte) error {
	v := struct {
		FnFindInMap [3]json.RawMessage `json:"Fn::FindInMap"`
	}{}
	if err := json.Unmarshal(buf, &v); err != nil {
		return err
	}
	if err := json.Unmarshal(v.FnFindInMap[0], &f.MapName); err != nil {
		return err
	}
	if err := json.Unmarshal(v.FnFindInMap[1], &f.TopLevelKey); err != nil {
		return err
	}
	if err := json.Unmarshal(v.FnFindInMap[2], &f.SecondLevelKey); err != nil {
		return err
	}

	return nil
}

func (f FindInMapFunc) String() *StringExpr {
	return &StringExpr{Func: f}
}

func (f FindInMapFunc) Exec(fm *Formation) (v interface{}, err error) {

	if mapping, ok := fm.Mappings[f.MapName]; ok {
		var topkey, seckey string = f.TopLevelKey.Literal, f.SecondLevelKey.Literal
		if f.TopLevelKey.Func != nil {
			tv, err := f.TopLevelKey.Func.Exec(fm)
			if err != nil {
				return "", errors.New("Exec FindInMapFunc Faid In Eval TopLevelKey. " + err.Error())
			}
			if topkey, ok = tv.(string); !ok {
				return "", errors.New("Exec FindInMapFunc Faid In Eval TopLevelKey. " + err.Error())
			}
		}
		if f.SecondLevelKey.Func != nil {
			sv, err := f.SecondLevelKey.Func.Exec(fm)
			if err != nil {
				return "", errors.New("Exec FindInMapFunc Faid In Eval SecondLevelKey. " + err.Error())
			}
			if seckey, ok = sv.(string); !ok {
				return "", errors.New("Exec FindInMapFunc Faid In Eval TopLevelKey. " + err.Error())
			}
		}
		tmap := map[string]map[string]string(*mapping)
		if smap, ok := tmap[topkey]; ok {
			if v, ok = smap[seckey]; ok {
				return v, nil
			} else {
				return "", errors.New("Exec FindInMapFunc Faid In Not Found SecondLevelKey[" + seckey + "].")
			}
		} else {
			return "", errors.New("Exec FindInMapFunc Faid In Not Found TopLevelKey[" + topkey + "].")
		}
	}
	return nil, errors.New("No Found Map[" + f.MapName + "]")

}
func (r FindInMapFunc) DependResource() []string {
	return nil
}

var _ Stringable = FindInMapFunc{} // FindInMapFunc must implement Stringable
var _ StringFunc = FindInMapFunc{} // FindInMapFunc must implement StringFunc
