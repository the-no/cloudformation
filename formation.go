package cloudformation

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/the-no/aws-sdk-go/aws"
	"github.com/the-no/aws-sdk-go/aws/session"
	"sync"
)

type Formation struct {
	Request                  *Request
	Platform                 Platform
	Session                  *session.Session
	AWSTemplateFormatVersion string
	Parameters               map[string]*ParameterValue
	Conditions               map[string]bool
	Mappings                 map[string]*Mapping
	Callback                 []string
	Resources                map[string]*ResourceUnit
	Outputs                  map[string]interface{}
}

func (self *Formation) StartResourceUnits() {
	var wg sync.WaitGroup
	for _, v := range self.Resources {
		wg.Add(1)
		go func() {
			defer wg.Done()
			startResourceUnit(self, v)
		}()
	}
}

func (self *Formation) Condition(c string) bool {

	if v, ok := self.Conditions[c]; ok {
		return v
	}
	return false
}

func (self *Formation) evalStructExpr(s json.RawMessage) ([]byte, error) {

	sdata := []byte(s)
	instring := false
	l := newLexer(sdata)
	//depend := []string{}
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
						return nil, err
					}
					val, err := f.Exec(self)
					if err != nil {
						return nil, err
					}
					vjson, err := json.Marshal(val)
					data = append(data, vjson...)
					//depend = append(depend, f.DependResource()...)
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
	return data, nil
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

	Input  interface{}
	Output interface{}
	Ref    aws.Referencer
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

func startResourceUnit(fm *Formation, r *ResourceUnit) error {
	if cond, ok := fm.Conditions[r.Resource.Condition]; ok && !cond {
		r.Err = errors.New("Condition [" + r.Resource.Condition + "] Is False OR Not Found In Create " + r.Name + "!")
		r.Done = true
		r.cond.Broadcast()
		return r.Err
	}

	for _, depend := range r.Resource.DependsOn {
		if res, ok := fm.Resources[depend]; ok {
			if err := res.Wait(); err != nil {
				r.Err = errors.New(r.Name + " Depend On [" + depend + "] " + err.Error() + "!")
				r.cond.Broadcast()
				return r.Err
			}
		}
	}

	data, err := fm.evalStructExpr(r.Resource.Properties)
	if err != nil {
		r.Err = errors.New("Eval Struct Expr Failed. " + err.Error())
		r.cond.Broadcast()
	}

	cli, err := fm.Platform.NewClinet("EC2", fm.Session)
	if err != nil {
		r.Err = errors.New("Create Request Clinet Failed. " + err.Error())
		r.cond.Broadcast()
	}

	r.Input, r.Output, r.Ref, err = cli.CreateResource(r.Resource.Type, data)
	if err != nil {
		r.Err = errors.New(r.Name + "Create Resource Failed. " + err.Error())
	}
	r.cond.Broadcast()
	fm.Callback = append(fm.Callback, r.Name)
	return nil
}
