package controllers

import (
	"github.com/jkoreska/gowebapi"
)

type testController struct {
	testme int64
	auther gowebapi.Auther
}

func NewTestController(auther gowebapi.Auther) *testController {
	return &testController{0, auther}
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

type AuthModel struct {
	Email string
	Secret string
}

func (self *testController) Authenticate(model *AuthModel, request *gowebapi.Request) *gowebapi.Response {

	if "test@test.com" == model.Email && "123123" == model.Secret {

		token := self.auther.Signin("test@test.com", 10)

		return &gowebapi.Response{
			Status: 200,
			Data: struct{Token string}{
				Token: token,
			},
		}
	}

	return &gowebapi.Response{
		Status: 401,
		Header: map[string][]string{"Www-Authenticate": []string{"Basic"}},
	}
}

type TestModel struct {
	Id     int64
	Tester interface{}
	User   string
}

func (self *testController) TestModel(id int, model *TestModel, request *gowebapi.Request) *gowebapi.Response {

	self.testme++
	model.Id = self.testme
	model.User = request.UserData

	return &gowebapi.Response{
		Status: 200,
		Data:   model,
	}
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
