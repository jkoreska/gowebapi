package controllers

import (
	"code.luktek.com/git/gowebapi"
)

type TestController struct {
	testme int64
}

func (self *TestController) TestRequest(request *gowebapi.Request) *gowebapi.Response {

	self.testme++

	return &gowebapi.Response{
		Status: 200,
		Data: TestModel{
			Id:     self.testme,
			Tester: request.Data["Tester"],
		},
	}
}

func (self *TestController) TestModel(id int, model *TestModel) *gowebapi.Response {

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

func (self *TestController) Get(id int64, test string) *gowebapi.Response {
	return &gowebapi.Response{
		Status: 200,
		Data:   &TestModel{Id: id, Tester: test},
	}
}

func (self *TestController) Post(model *TestModel) *gowebapi.Response {
	return &gowebapi.Response{
		Status: 201,
		Data:   model,
	}
}

func (self *TestController) Put(id int64, model *TestModel) *gowebapi.Response {
	return &gowebapi.Response{
		Status: 202,
		Data:   model,
	}
}

func (self *TestController) Delete(id int64) *gowebapi.Response {
	return &gowebapi.Response{
		Status: 210,
	}
}
