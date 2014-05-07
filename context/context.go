package context

import (
	"reflect"
	"encoding/json"
	"fmt"
)

type WorkloadContext interface{
	PutString(k string, v string)
	GetString(k string) string
	PutInt(k string, v int)
	GetInt(k string) int
	PutInt64(k string, v int64)
	GetInt64(k string) int64
	PutFloat64(k string, v float64)
	GetFloat64(k string) float64
	PutBool(k string, v bool)
	GetBool(k string) bool
	MarshalJSON() ([]byte, error)
	Unmarshal(data []byte) error
	CheckExists(k string) bool
	CheckType(k string) string
	Clone() WorkloadContent
	GetContent() WorkloadContent
	GetKeys() []string
}

type WorkloadContent struct{
	Content map[string]interface{}
}

func New() WorkloadContext {
	return WorkloadContext( WorkloadContent{make(map[string]interface{})} )
}

func NewWorkloadContent() WorkloadContent {
	return WorkloadContent{make(map[string]interface{})}
}

func (c WorkloadContent) PutString(k string, v string) {
	c.Content[k] = v
}

func (c WorkloadContent) GetString(k string) string {
	return c.Content[k].(string)
}

func (c WorkloadContent) PutInt(k string, v int) {
	c.Content[k] = v
}

func (c WorkloadContent) GetInt(k string) int {
	//json.Unmarsal put numbers as float64 into interface{}
	if reflect.TypeOf(c.Content[k]).Name() == "float64" {	
		c.Content[k] = int(c.Content[k].(float64))
	}
	return c.Content[k].(int)
}

func (c WorkloadContent) PutInt64(k string, v int64) {
	c.Content[k] = v
}

func (c WorkloadContent) GetInt64(k string) int64 {
	//json.Unmarsal put numbers as float64 into interface{}
	if reflect.TypeOf(c.Content[k]).Name() == "float64" {
		c.Content[k] = int64(c.Content[k].(float64))
	}
	return c.Content[k].(int64)
}

func (c WorkloadContent) PutFloat64(k string, v float64) {
	c.Content[k] = v
}

func (c WorkloadContent) GetFloat64(k string) float64 {
	return c.Content[k].(float64)
}

func (c WorkloadContent) PutBool(k string, v bool) {
	c.Content[k] = v
}

func (c WorkloadContent) GetBool(k string) bool {
	return c.Content[k].(bool)
}

func (c WorkloadContent) MarshalJSON() ([]byte, error) {
	fmt.Println("**** called marshal")
	return json.Marshal(c.Content)
}

func (c WorkloadContent) Unmarshal(data []byte) error {
	fmt.Println("**** called Unmarshal")
	return json.Unmarshal(data, &c.Content)	
}

func (c WorkloadContent) CheckExists(k string) bool {
	if c.Content[k] == nil {
		return false
	} else {
		return true
	}
}

func (c WorkloadContent) CheckType(k string) string {
	return reflect.TypeOf(c.Content[k]).Name()
}

func (c WorkloadContent) Clone() WorkloadContent {
	var clone = WorkloadContent{make(map[string]interface{})}
	for k, v := range c.Content {
    	clone.Content[k] = v
	}
	return clone
}

func (c WorkloadContent) GetContent() WorkloadContent {
	return c
}

func  (c WorkloadContent) GetKeys() []string {
	var keys []string
	for k, _ := range c.Content {
    	keys = append(keys, k)
	}
	return keys
}