package rest

import (
	"fmt"
	"net/http"
)

type Response struct {
	status      int         `json:"-"`
	Result      string      `json:"result"`
	Message     string      `json:"message,omitempty""`
	Content     interface{} `json:"content,omitempty"`
	contentType string      `json:"-"`
}

func (this *Response) HasError() bool {
	if this.status != http.StatusOK {
		return true
	}
	return false
}

func InvalidParameter(Message string) *Response {
	return &Response{
		status:  http.StatusBadRequest,
		Result:  "InvalidParameter",
		Message: Message,
	}
}

func NotLoginError() *Response {
	return &Response{
		status: http.StatusForbidden,
		Result: "UserNotLogin",
	}
}

func NotFoundError(what string, Message string) *Response {
	return &Response{
		status:  http.StatusNotFound,
		Result:  fmt.Sprintf("%sNotFound", what),
		Message: Message,
	}
}

func SystemError(what string, Message string) *Response {
	return &Response{
		status:  http.StatusServiceUnavailable,
		Result:  fmt.Sprintf("%sError", what),
		Message: Message,
	}
}

func PermissionDenied(who string, Message string) *Response {
	return &Response{
		status:  http.StatusForbidden,
		Result:  fmt.Sprintf("%sPermissionDenied", who),
		Message: Message,
	}
}

func Json(data interface{}) *Response {
	return &Response{
		status:      http.StatusOK,
		Result:      "ok",
		Content:     data,
		contentType: "application/json",
	}
}

func String(data string) *Response {
	return &Response{
		status:      http.StatusOK,
		Result:      "ok",
		Content:     data,
		contentType: "text/plain",
	}
}

func OK() *Response {
	return &Response{
		status:      http.StatusOK,
		Result:      "ok",
		Content:     "",
		contentType: "application/json",
	}
}
