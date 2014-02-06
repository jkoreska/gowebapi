package main

import (
	"fmt"
	"github.com/jkoreska/gowebapi"
	"github.com/jkoreska/gowebapi/sample/controllers"
	"net/http"
)

func main() {

	auther := gowebapi.NewDefaultAuther([]byte{0x6c, 0xf8, 0x05, 0x1b, 0x4a, 0xae, 0xc0, 0xa9, 0x7f, 0x47, 0x94, 0x8d, 0x11, 0xdf, 0xe0, 0x0a})
	testController := controllers.NewTestController(auther)

	handler := gowebapi.NewDefaultHandler()

	handler.Filter().
		Add(CorsFilter)

	handler.Router().
		AddRoute("/auth/").
		ForMethod("post").
		ToFunc(testController.Authenticate)

	handler.Router().
		AddRoute("/func/").
		ToMethod(testController, "TestModel").
		WithFilter(auther.Authenticate)

	handler.Router().
		AddRestRoutes("/rest/{id}/{test}", testController)

	//handler.Router().
	//	BindRpcRoutes("/rpc/", testController)

	fmt.Println("We're a GO!")

	http.ListenAndServe(":8888", handler)
}

func CorsFilter(request *gowebapi.Request) (*gowebapi.Response, bool) {

	response := &gowebapi.Response{
		Status: 200,
		Header: map[string][]string{"Access-Control-Allow-Origin": []string{"*"}},
	}

	next := "OPTIONS" != request.Http.Method

	return response, next
}
