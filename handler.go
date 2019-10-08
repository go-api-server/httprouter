package rest

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type HandlerFunc func(*Context, Parameter) *Response

type requestHandler struct {
	beforeFilterArray []FilterFunc
	afterFilterArray  []FilterFunc
	method            string
	uri               string
	handleFunc        interface{}
	handleFuncType    reflect.Type
	handleFuncParams  map[string]*parameter
}

func (this *requestHandler) doRequest(ctx *Context) *Response {
	for _, fn := range this.beforeFilterArray {
		res := fn(ctx)
		if res.HasError() {
			return res
		}
	}

	err := ctx.Request.ParseForm()
	if err != nil {
		return SystemError("ParseForm", err.Error())
	}

	param := reflect.TypeOf(this.handleFunc).In(1)
	arg := reflect.New(param)
	for i := 0; i < param.NumField(); i++ {
		tag := strings.Trim(param.Field(i).Tag.Get("web"), " ")
		if len(tag) == 0 {
			continue
		}
		arr := strings.Split(tag, ",")
		key := arr[0]
		str := ctx.Request.FormValue(key)
		if len(str) == 0 && len(arr) > 1 && arr[1] == "required" {
			return InvalidParameter(fmt.Sprintf("expect param: %s", key))
		}
		switch param.Field(i).Type.Kind() {
		case reflect.String:
			arg.Field(i).SetString(str)
		case reflect.Bool:
			val, _ := strconv.ParseBool(str)
			arg.Field(i).SetBool(val)
		case reflect.Float32, reflect.Float64:
			val, _ := strconv.ParseFloat(str, 64)
			arg.Field(i).SetFloat(val)
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			val, _ := strconv.ParseInt(str, 0, 64)
			arg.Field(i).SetInt(val)
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			val, _ := strconv.ParseUint(str, 0, 64)
			arg.Field(i).SetUint(val)
		}
	}

	res := this.handleFunc(ctx, arg.Interface().(Parameter))
	if res != nil {
		return res
	}

	for _, fn := range this.afterFilterArray {
		res := fn(ctx)
		if res.HasError() {
			return res
		}
	}

	return res
}
