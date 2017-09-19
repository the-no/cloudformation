package cloudformation

import (
	"encoding/json"
	"errors"
	"fmt"
	//"strings"
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
	depend := []string{}
	data := make([]byte, 0, len(sdata))

	for ch := l.peekChar(); ch != 0; ch = l.peekChar() {
		switch ch {
		case '{':
			if !instring {
				key := l.blockKey()
				if isFunc(key) {
					fjson := l.readBlock()
					fmt.Println(string(fjson))
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
					depend = append(depend, f.DependResource()...)
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
	if cond, ok := fm.Conditions[r.Resource.Condition]; ok && !cond {
		r.Err = errors.New("Condition [" + r.Resource.Condition + "] Is False OR Not Found In Create " + r.Name + "!")
		r.cond.Broadcast()
		return
	}

	for _, depend := range r.Resource.DependsOn {
		if res, ok := fm.Resources[depend]; ok {
			if err := res.Wait(); err != nil {
				r.Err = errors.New(r.Name + " Depend On [" + depend + "] " + err.Error() + "!")
				r.cond.Broadcast()
				return
			}
		}
	}

	data, depends, err := fm.evalStructExpr(r.Resource.Properties)
	fmt.Println("+++++++++++++++++", string(data), depends, err)
	return
}

func getplatform(resourcetype string) {

}
