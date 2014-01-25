package controllers

import (
	"github.com/jkoreska/gowebapi"
)

type testController struct {
	testme int64
}

func NewTestController() *testController {
	return &testController{}
}

func (self *testController) TestRequest(request *gowebapi.Request) *gowebapi.Response {

	self.testme++

	return &gowebapi.Response{
		Status: 200,
		Data: TestModel{
			Id:     self.testme,
			Tester: request.Data["Tester"],
		},
	}
}

func (self *testController) TestModel(id int, model *TestModel) *gowebapi.Response {

	self.testme++
	model.Id = self.testme

	return &gowebapi.Response{
		Status: 201,
		Data:   model,
	}
}

type TestModel struct {
	Id     int64
	Tester interface{}
}

func (self *testController) Get(id int64, test string) *gowebapi.Response {
	return &gowebapi.Response{
		Status: 200,
		Data:   &TestModel{Id: id, Tester: test},
	}
}

func (self *testController) Post(model *TestModel) *gowebapi.Response {
	return &gowebapi.Response{
		Status: 201,
		Data:   model,
	}
}

func (self *testController) Put(id int64, model *TestModel) *gowebapi.Response {
	return &gowebapi.Response{
		Status: 202,
		Data:   model,
	}
}

func (self *testController) Delete(id int64) *gowebapi.Response {
	return &gowebapi.Response{
		Status: 210,
	}
}
