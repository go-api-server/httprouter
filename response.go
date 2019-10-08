package rest

import (
	"fmt"
	"net/http"
)

type Response struct {
	status      int
	result      string
	message     string
	content     interface{}
	contentType string
}

func (this *Response) HasError() bool {
	if this.status != http.StatusOK {
		return true
	}
	return false
}

func InvalidParameter(message string) *Response {
	return &Response{
		status:  http.StatusBadRequest,
		result:  "InvalidParameter",
		message: message,
	}
}

func NotLoginError() *Response {
	return &Response{
		status: http.StatusForbidden,
		result: "UserNotLogin",
	}
}

func NotFoundError(what string, message string) *Response {
	return &Response{
		status:  http.StatusNotFound,
		result:  fmt.Sprintf("%sNotFound", what),
		message: message,
	}
}

func SystemError(what string, message string) *Response {
	return &Response{
		status:  http.StatusServiceUnavailable,
		result:  fmt.Sprintf("%sNotFound", what),
		message: message,
	}
}

func PermissionDenied(who string, message string) *Response {
	return &Response{
		status:  http.StatusForbidden,
		result:  fmt.Sprintf("%sPermissionDenied", who),
		message: message,
	}
}

func JsonResponse(status int, data interface{}) *Response {
	return &Response{
		status:      http.StatusOK,
		result:      "ok",
		content:     data,
		contentType: "application/json",
	}
}

func StringResponse(status int, data string) *Response {
	return &Response{
		status:      http.StatusOK,
		result:      "ok",
		content:     data,
		contentType: "text/plain",
	}
}
