package context

import (
	"encoding/json"
	"strconv"
)

type Context interface {
	PutString(k string, v string)
	GetString(k string) (string, bool)
	PutInt(k string, v int)
	GetInt(k string) (int, bool)
	PutFloat64(k string, v float64)
	GetFloat64(k string) (float64, bool)
	PutBool(k string, v bool)
	GetBool(k string) (bool, bool)
	MarshalJSON() ([]byte, error)
	UnmarshalJSON(b []byte) error
	Clone() contextMap
}

type contextMap map[string]interface{}

func New() Context {
	var c contextMap = make(contextMap)
	return &c
}

func (c contextMap) PutString(k string, v string) {
	c[k] = v
}

func (c contextMap) GetString(k string) (string, bool) {
	if c[k] == nil {
		return "", false
	} else {
		return c[k].(string), true
	}
}

func (c contextMap) PutInt(k string, v int) {
	// saves as string, avoids json.Marshal turning int into json.numbers(float64)
	c[k] = strconv.Itoa(v)
}

func (c contextMap) GetInt(k string) (int, bool) {
	if c[k] == nil {
		return 0, false
	} else {
		value, _ := strconv.ParseInt(c[k].(string), 0, 0)
		return int(value), true
	}
}

func (c contextMap) PutFloat64(k string, v float64) {
	c[k] = v
}

func (c contextMap) GetFloat64(k string) (float64, bool) {
	if c[k] == nil {
		return 0, false
	} else {
		return c[k].(float64), true
	}
}

func (c contextMap) PutBool(k string, v bool) {
	c[k] = v
}

func (c contextMap) GetBool(k string) (bool, bool) {
	if c[k] == nil {
		return false, false
	} else {
		return c[k].(bool), true
	}
}

func (c contextMap) MarshalJSON() ([]byte, error) {
	return json.Marshal(map[string]interface{}(c))
}

func (c contextMap) UnmarshalJSON(b []byte) error {
	var m map[string]interface{} = c
	return json.Unmarshal(b, &m)
}

func (c contextMap) Clone() contextMap {
	var clone = make(contextMap)
	for k, v := range c {
		clone[k] = v
	}
	return clone
}
