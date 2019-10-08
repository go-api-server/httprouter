package main

import (
	"github.com/golib-svr/rest"
)

type query struct{}

func Home(ctx *rest.Context, param query) *rest.Response {
	return nil
}

func main() {
	r := rest.NewRouter()
	r.Get("/", Home)
}
