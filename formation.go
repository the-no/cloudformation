package cloudformation

import (
	"encoding/json"
	"fmt"
	//"io/ioutil"
	"errors"
	//	"reflect"
	//"regexp"
	"strings"
	"sync"
)

type Formation struct {
	Request                  *Request
	AWSTemplateFormatVersion string
	Parameters               map[string]*ParameterValue
	Conditions               map[string]bool
	Mappings                 map[string]*Mapping

	Resources map[string]*ResourceUnit
	Outputs   map[string]interface{}
}

func (self *Formation) StartResourceUnits() {
	for _, v := range self.Resources {
		go startResourceUnit(self, v)
		break
	}

}

/*func (self *Formation) callFunc(f Func) (interface{}, error) {

	switch f.(type) {
	case RefFunc:
	case JoinFunc:
	case EqualsFunc:
	case FindInMapFunc:
		findinmap := f.(FindInMapFunc)
		return self.findinmapFunc(findinmap)
	default:
	}
	return nil, nil
}
*/
/*func (self *Formation) funcDepend(f Func) []*ResourceUnit {

	depend := []*ResourceUnit{}
	switch f.(type) {
	case JoinFunc:
	case EqualsFunc:
		equals := f.(EqualsFunc)
		if equals.Value1.Func != nil {
			dps := self.funcDepend(equals.Value1.Func)
			depend = append(depend, dps...)
		}
		if equals.Value2.Func != nil {
			dps := self.funcDepend(equals.Value2.Func)
			depend = append(depend, dps...)
		}
	case FindInMapFunc:
		findinmap := f.(FindInMapFunc)
		if findinmap.TopLevelKey.Func != nil {
			dps := self.funcDepend(findinmap.TopLevelKey.Func)
			depend = append(depend, dps...)
		}
		if findinmap.SecondLevelKey.Func != nil {
			dps := self.funcDepend(findinmap.SecondLevelKey)
			depend = append(depend, dps...)
		}
	default:
	}
	return nil, nil
}*/

/*func (self *Formation) refFunc(f RefFunc) (string, error) {
	if v, ok := self.Parameters[f.Name]; ok {
		return v.Value, nil
	}

	if v, ok := self.Resources[f.Name]; ok {
		if err := v.Wait(); err != nil {
			return "", err
		}
		return "resval", nil
	}
	return self.PseudoParameter(f.Name)
}*/

/*func (self *Formation) equalsFunc(f EqualsFunc) (bool, error) {

}

func (self *Formation) findinmapFunc(f FindInMapFunc) (string, error) {

	topkey, err := self.evalStringExpr(&f.TopLevelKey)
	if err != nil {
		return "", err
	}

	seckey, err := self.evalStringExpr(&f.SecondLevelKey)
	if err != nil {
		return "", err
	}
	if mapping, ok := self.Mappings[f.MapName]; ok {
		topmap := map[string]map[string]string(*mapping)
		if secmap, ok := topmap[topkey]; ok {
			if v, ok := secmap[seckey]; ok {
				return v, nil
			}
			return "", errors.New(seckey + " No Found In Second Map[" + f.MapName + "]!")
		}
		return "", errors.New(topkey + " No Found In Top Map[" + f.MapName + "]!")
	}
	return "", errors.New("Map[" + f.MapName + "]  No Found!")
}
*/
/*func (self *Formation) evalStringExpr(expr *StringExpr) (string, error) {
	if expr.Func != nil {
		result, err := expr.Func.Exec(self)
		if err != nil {
			return "", err
		}
		return result.(string), nil
	}
	return expr.Literal, nil
}*/

/*func (self *Formation) evalIntegerExpr(expr *IntegerExpr) (int64, error) {
	if expr.Func != nil {
		result, err := expr.Func.Exec(self)
		if err != nil {
			return 0, err
		}
		return result.(int64), nil
	}
	return expr.Literal, nil
}
*/
/*func (self *Formation) evalBoolExpr(expr *BoolExpr) (bool, error) {

	if expr.Func != nil {
		result, err := expr.Func.Exec(self)
		if err != nil {
			return false, err
		}
		return result.(bool), nil
	}
	return expr.Literal, nil
}
*/
func (self *Formation) Condition(c string) bool {

	if v, ok := self.Conditions[c]; ok {
		return v
	}
	return false
}

func (self *Formation) evalStructExpr(s json.RawMessage) ([]byte, []*ResourceUnit, error) {

	sdata := []byte(s)
	instring := false
	l := newLexer(sdata)
	depend := []*ResourceUnit{}
	data := make([]byte, 0, len(sdata))

	for ch := l.peekChar(); ch != 0; ch = l.peekChar() {
		switch ch {
		case '{':
			if !instring {
				key := l.blockKey()
				if isFunc(key) {
					fjson := l.readBlock()
					f, err := unmarshalFunc(fjson)
					if err != nil {
						return nil, nil, err
					}
					val, err := f.Exec(self)
					if err != nil {
						return nil, nil, err
					}
					vjson, err := json.Marshal(val)
					data = append(data, vjson...)
					depend = append(depend, f.DependResource(self)...)
				} else {
					data = append(data, ch)
				}
			} else {
				data = append(data, ch)
			}
		case '"':
			instring = !instring
			data = append(data, ch)
		case '\\':
			data = append(data, ch)
			l.readChar()
			data = append(data, l.peekChar())
		default:
			data = append(data, ch)
		}
		l.readChar()
	}

	fmt.Println("struct :", string(data))
	return data, nil, nil
}

func (self *Formation) PseudoParameter(name string) (string, error) {

	switch name {
	case "AWS::Region":
		return "us-west-2", nil
	case "AWS::StackId":
	case "AWS::StackName":
		return self.Request.StackName, nil
	case "AWS::NoValue":
		return "", nil
		//	case "AWS::NotificationARNs":
	case "AWS::AccountId":
	}
	return "", errors.New("Invail Parameter")
}

type ParameterValue struct {
	Type  string
	Name  string
	Value string
}

type ResourceUnit struct {
	Fm       *Formation
	Resource *Resource
	Name     string
	Done     bool
	Err      error
	Result   interface{}
	Callback bool
	cond     *sync.Cond
}

func NewResourceUnit(fm *Formation, name string, r *Resource) *ResourceUnit {
	return &ResourceUnit{
		Fm:       fm,
		Name:     name,
		Resource: r,
		cond:     sync.NewCond(&sync.Mutex{}),
	}
}

func (self *ResourceUnit) Wait() error {
	self.cond.L.Lock()
	if !self.Done {
		self.cond.Wait()
	}
	self.cond.L.Unlock()
	return self.Err
}

func startResourceUnit(fm *Formation, r *ResourceUnit) {
	if cond, ok := fm.Conditions[r.Resource.Condition]; !cond || !ok {
		r.Err = errors.New("Condition [" + r.Resource.Condition + "] Is False OR Not Found In Create " + r.Name + "!")
		r.cond.Broadcast()
		return
	}

	for _, depend := range r.Resource.DependsOn {
		if err := r.Wait(); err != nil {
			r.Err = errors.New(r.Name + " Depend On [" + depend + "] " + err.Error() + "!")
			r.cond.Broadcast()
			return
		}
	}
	data, _, err := fm.evalStructExpr(r.Resource.Properties)
	fmt.Println(string(data), err)
	return
}

func getplatform(resourcetype string) {

}

//func getResourceCreateParam(platform, product, resource string) {
func getResourceCreateParam(resourcetype string) (interface{}, error) {

	types := strings.Split(resourcetype, "::")
	switch types[0] {
	case "ALIYUN":
	case "AWS":
		getAWSResourceCreateParam(types[1], types[2])
	case "":
	}
	return nil, errors.New("Invalid Platform [" + types[0] + "]")
}
