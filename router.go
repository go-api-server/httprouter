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

var (
	uriHandlerMapper         map[string]*requestHandler
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
	})
}

func getRequestKey(method string, uri string, uriPrefix string) string {
	if len(uriPrefix) > 0 {
		return fmt.Sprintf("%s %s/%s", strings.ToLower(method), uriPrefix, strings.ToLower(uri))
	}
	return fmt.Sprintf("%s %s", strings.ToLower(method), strings.ToLower(uri))
}

func getRequestHandler(method string, uri string) *requestHandler {
	key := getRequestKey(method, uri, "")
	hdl, ok := uriHandlerMapper[key]
	if !ok {
		return nil
	}
	return hdl
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
		RequestID: fmt.Sprintf("%d_%d", time.Now().UnixNano(), rand.Intn(1000000)),
		Request:   r,
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

	key := getRequestKey(method, this.uriPrefix, uri)
	hdl := &requestHandler{
		beforeFilterArray: this.beforeFilterArray,
		afterFilterArray:  this.afterFilterArray,
		method:            method,
		uri:               uri,
		handleFunc:        fn,
		handleFuncType:    typeof,
		handleFuncParams:  make(map[string]*parameter),
	}

	for i := 0; i < typeof.NumIn(); i++ {
		kind := typeof.In(i).Kind()
		if kind == reflect.Func || kind == reflect.Map || kind == reflect.Slice {
			continue
		}

		if kind == reflect.Struct {
			for j := 0; j < typeof.In(i).NumField(); j++ {
				tag := strings.Trim(typeof.In(i).Field(j).Tag.Get("web"), " ")
				if len(tag) == 0 {
					continue
				}
				arr := strings.Split(tag, ",")
				notNull := false
				if len(arr) > 1 && strings.Trim(arr[1]) == "required" {
					notNull = true
				}
				name := strings.Trim(arr[0])
				hdl.handleFuncParams[name] = &parameter{
					NotNull: notNull,
					Type:    typeof.In(i).Field(j).Type,
				}
			}
			continue
		}

		hdl.handleFuncParams[typeof.In(i).Name()] = &parameter{
			NotNull: true,
			Type:    typeof.In(i),
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
		data, err = json.Marshal(r.content)
		if err != nil {
			data = []byte(err.Error())
		}

	} else {
		data = []byte(r.content.(string))
	}

	w.WriteHeader(r.status)
	w.Header().Set("Content-Type", r.contentType)
	w.Write(data)
}
