package rest

type FilterFunc func(*Context) *Response

type Filter interface {
	BeforeRequest(*Context) *Response
	AfterRequest(*Context) *Response
}
