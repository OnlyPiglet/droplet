package main

import (
	"log"
	"reflect"

	"github.com/OnlyPiglet/droplet/core"
	"github.com/OnlyPiglet/droplet/wrapper"
	fasthttpwrap "github.com/OnlyPiglet/droplet/wrapper/fasthttp"
	"github.com/fasthttp/router"
	"github.com/valyala/fasthttp"
)

func main() {
	r := router.New()
	r.POST("/json_input/{id}", fasthttpwrap.Wraps(JsonInputDo,
		wrapper.InputType(reflect.TypeOf(&JsonInput{}))))
	log.Fatal(fasthttp.ListenAndServe(":8080", r.Handler))
}

type JsonInput struct {
	ID    string   `auto_read:"id,path" json:"id"`
	User  string   `auto_read:"user,header" json:"user"`
	IPs   []string `json:"ips"`
	Count int      `json:"count"`
	Body  []byte   `auto_read:"@body"`
}

func JsonInputDo(ctx core.Context) (interface{}, error) {
	input := ctx.Input().(*JsonInput)

	return input, nil
}
