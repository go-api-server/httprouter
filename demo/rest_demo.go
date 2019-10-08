package main

import (
	"fmt"

	"github.com/golib-svr/rest"
)

type query struct {
	Id  int64 `form:"id"`
	Age int64 `form:"age"`
}

type body struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

func Home(ctx *rest.Context, param query) *rest.Response {
	fmt.Println("param.Id: ", param.Id)
	fmt.Println("param.Age: ", param.Age)
	return rest.Json(param)
}

func Say(ctx *rest.Context, data body) *rest.Response {
	return rest.Json(data)
}

func main() {
	r := rest.NewRouter()
	r.Get("/", Home)
	r.Post("/say", Say)
	rest.Serve(":88", r)
}
