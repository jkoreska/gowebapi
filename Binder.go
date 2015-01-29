package gowebapi

import (
	"errors"
	"reflect"
	"strconv"
	"strings"
	"encoding/json"
)

type Binder interface {
	Bind(*Request) (*Response, error)
}

type DefaultBinder struct{}

func (self *DefaultBinder) Bind(request *Request) (*Response, error) {

	if nil == request.Route.Target {
		return &Response{
			Status: 500,
			Data:   "Route has no target",
		}, nil
	}

	switch request.Route.Target.(type) {
	case func(*Request) *Response:
		return request.Route.Target.(func(*Request) *Response)(request), nil
	default:
		return self.bindWithReflect(request)
	}
}

func (self *DefaultBinder) bindWithReflect(request *Request) (*Response, error) {

	target := reflect.ValueOf(request.Route.Target)

	if reflect.Ptr == target.Kind() && reflect.Struct == target.Elem().Kind() {
		target = target.MethodByName(request.Route.Action)
		// handle invalid action
	}

	if reflect.Func != target.Kind() {
		return nil, errors.New("Invalid route target (expecting func)")
	}

	args, bindErr := self.bindArgs(target.Type(), request)

	if nil != bindErr {
		return nil, errors.New("Binding error: " + bindErr.Error())
	}

	retvals := target.Call(args)

	if 1 != len(retvals) || "*gowebapi.Response" != retvals[0].Type().String() {
		return nil, errors.New("Invalid route target response (expecting *gowebapi.Response)")
	}

	response := retvals[0].Interface().(*Response)

	if nil == response {
		return nil, errors.New("Invalid route target response (response is empty)")
	}

	return response, nil
}

func (self *DefaultBinder) bindArgs(targetType reflect.Type, request *Request) ([]reflect.Value, error) {

	args := make([]reflect.Value, 0)
	var err error

	for argNum := 0; argNum < targetType.NumIn(); argNum++ {

		argType := targetType.In(argNum)
		var arg reflect.Value

		if reflect.Struct == argType.Kind() ||
			(reflect.Ptr == argType.Kind() && reflect.Struct == argType.Elem().Kind()) {

			if "*gowebapi.Request" == argType.String() {

				args = append(args, reflect.ValueOf(request))

			} else {

				arg, err = self.bindStruct(argType, request.Data)

				if nil != err {
					continue
				}

				args = append(args, arg)
			}
		} else {

			if argNum < request.Route.Path.NumSubexp() {

				paramName := request.Route.Path.SubexpNames()[argNum+1]
				paramValue := request.Route.Params[paramName]

				arg, err = self.bindParam(argType, paramValue)

				if nil != err {
					continue
				}

				args = append(args, arg)

			} else {

				args = append(args, reflect.New(argType).Elem())
			}
		}
	}

	return args, err
}

func (self *DefaultBinder) bindParam(argType reflect.Type, paramValue string) (reflect.Value, error) {

	param := reflect.New(argType).Elem()

	switch argType.Kind() {
		case reflect.Bool:
			value, _ := strconv.ParseBool(paramValue)
			param.Set(reflect.ValueOf(value))
		case reflect.Int64:
			value, _ := strconv.ParseInt(paramValue, 10, 0)
			param.Set(reflect.ValueOf(value))
		case reflect.Float64:
			value, _ := strconv.ParseFloat(paramValue, 0)
			param.Set(reflect.ValueOf(value))
		default:
			param.Set(reflect.ValueOf(paramValue))
	}

	return param, nil
}

func (self *DefaultBinder) bindStruct(structType reflect.Type, data map[string]interface{}) (reflect.Value, error) {

	if reflect.Ptr == structType.Kind() {
		structType = structType.Elem()
	}

	arg := reflect.New(structType)

	for fieldNum := 0; fieldNum < arg.Elem().NumField(); fieldNum++ {

		field := arg.Elem().Field(fieldNum)
		structField := structType.Field(fieldNum)

		fieldName := structField.Name
		// use json tag field name if defined
		if ("" != structField.Tag.Get("json")) {
			fieldName = strings.Split(structField.Tag.Get("json"), ",")[0]
		}

		if nil != data[fieldName] {
			field.Set(self.bindField(structField.Type, data[fieldName]))
		}
	}

	return arg, nil
}

func (self *DefaultBinder) bindField(fieldType reflect.Type, data interface{}) (reflect.Value) {

	dataType := reflect.TypeOf(data)

	// TODO: validation goes here

	if reflect.Map == dataType.Kind() {

		fieldValue, _ := self.bindStruct(fieldType, data.(map[string]interface{}))

		return fieldValue
	
	} else if reflect.Array == dataType.Kind() || reflect.Slice == dataType.Kind() {
	
		arrayType := fieldType.Elem()
		fieldArray := reflect.MakeSlice(fieldType, 0, 10)

		for _, element := range data.([]interface{}) {

			fieldValue := self.bindField(arrayType, element)
			fieldArray = reflect.Append(fieldArray, fieldValue)
		}

		return fieldArray
	
	} else if reflect.PtrTo(fieldType).Implements(reflect.TypeOf((*json.Unmarshaler)(nil)).Elem()) {

		fieldValue := reflect.New(fieldType)
		fieldValue.
			MethodByName("UnmarshalJSON").
			Call(
				[]reflect.Value{
					reflect.ValueOf( []byte("\"" + data.(string) + "\"") ),
				},
			)

		return fieldValue.Elem()

	} else if dataType.AssignableTo(fieldType) {
	
		return reflect.ValueOf(data)

	} else if dataType.ConvertibleTo(fieldType) {
	
		return reflect.ValueOf(data).Convert(fieldType)
	}

	return reflect.New(fieldType)
}
