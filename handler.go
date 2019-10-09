package rest

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

type requestHandler struct {
	beforeFilterArray   []FilterFunc
	afterFilterArray    []FilterFunc
	method              string
	uri                 string
	uriPart             map[int]string
	handleFunc          reflect.Value
	handleFuncType      reflect.Type
	handleFuncParamType reflect.Type
	parameterMapper     map[string]*parameter
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

	argv := make([]reflect.Value, 0, 2)
	argv = append(argv, reflect.ValueOf(ctx))

	isPostJson := false
	if strings.Contains(ctx.Request.Header.Get("Content-Type"), "json") {
		isPostJson = true
	}

	uriParameterMapper := make(map[string]string)
	if len(this.uriPart) > 0 {
		uri := strings.Split(ctx.Request.RequestURI, "?")
		arr := strings.Split(uri[0], "/")
		cnt := len(arr)
		for i, k := range this.uriPart {
			if i >= cnt {
				continue
			}
			uriParameterMapper[k] = arr[i]
		}
	}

	if len(this.parameterMapper) > 0 {
		arg := reflect.New(this.handleFuncParamType)
		arg = arg.Elem()

		for name, param := range this.parameterMapper {
			str, ok := uriParameterMapper[name]
			if !ok {
				str = ctx.Request.FormValue(name)
			}

			if len(str) == 0 && param.NotNull {
				if isPostJson {
					continue
				}
				return InvalidParameter(fmt.Sprintf("expect %s param", name))
			}

			field := arg.FieldByName(param.Field)
			switch param.Type.Kind() {
			case reflect.String:
				field.SetString(str)
			case reflect.Bool:
				val, err := strconv.ParseBool(str)
				if err == nil {
					field.SetBool(val)
				}
			case reflect.Float32, reflect.Float64:
				val, err := strconv.ParseFloat(str, 64)
				if err == nil {
					field.SetFloat(val)
				}
			case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
				val, err := strconv.ParseInt(str, 0, 64)
				if err == nil {
					field.SetInt(val)
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				val, err := strconv.ParseUint(str, 0, 64)
				if err == nil {
					field.SetUint(val)
				}
			}
		}

		if isPostJson {
			body := make([]byte, ctx.Request.ContentLength)
			ctx.Request.Body.Read(body)
			err := json.Unmarshal(body, arg.Interface())
			if err != nil {
				fmt.Println(err.Error())
			}
		}

		argv = append(argv, arg)

	} else if isPostJson {
		arg := reflect.New(this.handleFuncParamType)

		body := make([]byte, ctx.Request.ContentLength)
		ctx.Request.Body.Read(body)
		err := json.Unmarshal(body, arg.Interface())
		if err != nil {
			fmt.Println(err.Error())
		}

		argv = append(argv, arg.Elem())

	} else if ctx.Request.ContentLength > 0 {
		arg := reflect.New(this.handleFuncParamType)
		arg = arg.Elem()
		argv = append(argv, arg)
	}

	var result *Response

	resultArray := this.handleFunc.Call(argv)
	resultCount := len(resultArray)
	if resultCount == 0 {
		result = OK()
	} else {
		if resultArray[0].IsNil() {
			result = OK()
		} else {
			val := resultArray[0].Interface()
			switch val.(type) {
			case *Response:
				result = val.(*Response)
			case error:
				err := val.(error)
				result = SystemError("", err.Error())
			}
		}
	}

	for _, fn := range this.afterFilterArray {
		res := fn(ctx)
		if res.HasError() {
			return res
		}
	}

	return result
}
