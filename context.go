package rest

import (
	"net/http"
)

type Context struct {
	RequestID string
	Request   *http.Request
	data      map[string]interface{}
}

func (this *Context) Get(key string) (interface{}, bool) {
	res, ok := this.data[key]
	if !ok {
		return nil, false
	}
	return res, true
}

func (this *Context) Set(key string, value interface{}) {
	this.data[key] = value
}
