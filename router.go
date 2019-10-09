package rest

import (
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"reflect"
	"runtime/debug"
	"strings"
	"sync"
	"time"
)

type uriPart struct {
	mapper map[string][]int
}

var (
	uriHandlerMapper         map[string]*requestHandler
	uriPartMapper            map[int]*uriPart
	initUriHandlerMapperOnce sync.Once
)

type router struct {
	uriPrefix         string
	beforeFilterArray []FilterFunc
	afterFilterArray  []FilterFunc
}

func NewRouter() *router {
	initRouter()
	return &router{
		beforeFilterArray: make([]FilterFunc, 0),
		afterFilterArray:  make([]FilterFunc, 0),
	}
}

func NewSubRouter(uriPrefix string) *router {
	initRouter()
	return &router{
		uriPrefix:         uriPrefix,
		beforeFilterArray: make([]FilterFunc, 0),
		afterFilterArray:  make([]FilterFunc, 0),
	}
}

func initRouter() {
	initUriHandlerMapperOnce.Do(func() {
		uriHandlerMapper = make(map[string]*requestHandler)
		uriPartMapper = make(map[int]*uriPart)
	})
}

func getRequestKey(method string, uri string, uriPrefix string) string {
	if len(uriPrefix) > 0 {
		return fmt.Sprintf("%s %s/%s", strings.ToLower(method), uriPrefix, strings.ToLower(uri))
	}
	return fmt.Sprintf("%s %s", strings.ToLower(method), strings.ToLower(uri))
}

func getRequestHandler(method string, uri string) *requestHandler {
	arr := strings.Split(uri, "?")
	key := getRequestKey(method, arr[0], "")

	hdl, ok := uriHandlerMapper[key]
	if ok {
		return hdl
	}

	arr = strings.Split(key, "/")
	cnt := len(arr)
	part, ok := uriPartMapper[cnt]
	if !ok {
		fmt.Println("uri len not match")
		return nil
	}

	for key, idx := range part.mapper {
		for _, i := range idx {
			arr[i] = ":"
		}
		uri := strings.Join(arr, "/")
		if key == uri {
			hdl, ok := uriHandlerMapper[key]
			if ok {
				return hdl
			}
			break
		}
	}

	return nil
}

// ServeHTTP for http.Handler interface
func (this *router) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			debug.PrintStack()
			log.Println(err)
		}
	}()

	ctx := &Context{
		Request:   r,
		RequestID: fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(1000000)),
		data:      make(map[string]interface{}),
	}

	hdl := getRequestHandler(r.Method, r.RequestURI)
	if hdl == nil {
		response(w, NotFoundError("URL", ""))
		return
	}

	res := hdl.doRequest(ctx)
	if res.HasError() {
		response(w, res)
		return
	}

	response(w, res)
}

func (this *router) AppendFilterBeforeRequest(fn FilterFunc) {
	this.beforeFilterArray = append(this.beforeFilterArray, fn)
}

func (this *router) AppendFilterAfterRequest(fn FilterFunc) {
	this.afterFilterArray = append(this.afterFilterArray, fn)
}

func (this *router) AppendFilter(filter Filter) {
	this.beforeFilterArray = append(this.beforeFilterArray, filter.BeforeRequest)
	this.afterFilterArray = append(this.afterFilterArray, filter.AfterRequest)
}

func (this *router) SetRequestHandler(method string, uri string, fn interface{}) {
	typeof := reflect.TypeOf(fn)

	if typeof.Kind() != reflect.Func {
		fmt.Printf("%s is not a func\n", typeof.Name())
		return
	}

	if typeof.NumIn() < 1 {
		fmt.Printf("%s need one parameter least\n", typeof.Name())
		return
	}

	if typeof.In(0).Kind() != reflect.Ptr {
		fmt.Printf("%s first parameter must be *Context\n", typeof.Name())
		return
	}

	first := typeof.In(0).Elem()
	if first.Name() != reflect.TypeOf(Context{}).Name() {
		fmt.Printf("%s first parameter must be *Context\n", typeof.Name())
		return
	}

	if typeof.NumIn() < 2 {
		return
	}

	second := typeof.In(1)
	if second.Kind() != reflect.Struct {
		fmt.Printf("%s second parameter must be struct\n", typeof.Name())
		return
	}

	key := ""
	hdl := &requestHandler{
		beforeFilterArray:   this.beforeFilterArray,
		afterFilterArray:    this.afterFilterArray,
		method:              method,
		uri:                 uri,
		uriPart:             make(map[int]string),
		handleFunc:          reflect.ValueOf(fn),
		handleFuncType:      typeof,
		handleFuncParamType: second,
		parameterMapper:     make(map[string]*parameter),
	}

	if strings.Contains(uri, ":") {
		arr := strings.Split(uri, "/")
		cnt := len(arr)
		idx := make([]int, 0, 1)
		res := make([]string, cnt)
		for i, v := range arr {
			part := strings.Split(v, ":")
			if len(part) > 1 {
				hdl.uriPart[i] = part[1]
				idx = append(idx, i)
				res[i] = ":"
			} else {
				res[i] = part[0]
			}
		}

		key = getRequestKey(method, strings.Join(res, "/"), this.uriPrefix)
		m, ok := uriPartMapper[cnt]
		if ok {
			_, ok := m.mapper[key]
			if ok {
				return
			}
			m.mapper[key] = idx
		} else {
			m = &uriPart{
				mapper: make(map[string][]int),
			}
			m.mapper[key] = idx
			uriPartMapper[cnt] = m
		}

	} else {
		key = getRequestKey(method, uri, this.uriPrefix)
	}

	// 处理参数
	for i := 0; i < second.NumField(); i++ {
		kind := second.Field(i).Type.Kind()
		if kind == reflect.Ptr || kind == reflect.Func || kind == reflect.Map || kind == reflect.Slice {
			continue
		}

		tag := strings.Trim(second.Field(i).Tag.Get("form"), " ")
		if len(tag) == 0 {
			continue
		}

		arr := strings.Split(tag, ",")
		notNull := false
		if len(arr) > 1 && strings.Trim(arr[1], " ") == "notnull" {
			notNull = true
		}

		name := strings.Trim(arr[0], " ")
		hdl.parameterMapper[name] = &parameter{
			NotNull: notNull,
			Field:   second.Field(i).Name,
			Type:    second.Field(i).Type,
		}
	}

	uriHandlerMapper[key] = hdl
}

func (this *router) Get(uri string, fn interface{}) {
	this.SetRequestHandler(http.MethodGet, uri, fn)
}

func (this *router) Put(uri string, fn interface{}) {
	this.SetRequestHandler(http.MethodPut, uri, fn)
}

func (this *router) Post(uri string, fn interface{}) {
	this.SetRequestHandler(http.MethodPost, uri, fn)
}

func (this *router) Delete(uri string, fn interface{}) {
	this.SetRequestHandler(http.MethodDelete, uri, fn)
}

func (this *router) Options(uri string, fn interface{}) {
	this.SetRequestHandler(http.MethodOptions, uri, fn)
}

func response(w http.ResponseWriter, r *Response) {
	if r.status == 0 {
		r.status = http.StatusOK
	}

	var err error
	var data []byte
	if len(r.contentType) == 0 || r.contentType == "application/json" {
		r.contentType = "application/json"
		data, err = json.Marshal(r)
		if err != nil {
			data = []byte(err.Error())
			fmt.Println("err: ", err.Error())
		}

	} else {
		data = []byte(r.Content.(string))
	}

	w.WriteHeader(r.status)
	w.Header().Set("Content-Type", r.contentType)
	w.Write(data)
}

func Serve(addr string, r *router) {
	http.ListenAndServe(addr, r)
}
