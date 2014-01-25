package main

import (
	"fmt"
	"github.com/jkoreska/gowebapi"
	"github.com/jkoreska/gowebapi/sample/controllers"
	"net/http"
)

func main() {

	testController := controllers.NewTestController()

	handler := gowebapi.NewDefaultHandler()

	handler.Router().
		AddRoute("/func/").
		ForMethod("post").
		ToFunc(testController.TestRequest)

	handler.Router().
		AddRoute("/func/").
		ToMethod(testController, "TestModel")

	handler.Router().
		AddRestRoutes("/rest/{id}/{test}", testController)

	//handler.Router().
	//	BindRpcRoutes("/rpc/", testController)

	fmt.Println("We're a GO!")

	http.ListenAndServe(":8888", handler)
}
