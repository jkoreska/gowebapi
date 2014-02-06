package gowebapi

import (
	"reflect"
	"regexp"
)

type Route struct {
	Path    *regexp.Regexp
	Method  *regexp.Regexp
	Headers map[string]*regexp.Regexp
	Params  map[string]string
	Target  interface{}
	Action  string
	Binder
	Filter *Filter
}

func (self *Route) ForMethod(method string) *Route {

	self.Method = regexp.MustCompile("(?i)" + method)

	return self
}

func (self *Route) ForHeader(name string, value string) *Route {

	self.Headers[name] = regexp.MustCompile(value)

	return self
}

func (self *Route) ToFunc(target interface{}) *Route {

	targetType := reflect.TypeOf(target)

	if reflect.Func != targetType.Kind() {

		panic("Invalid target type (expecting func)")

	} else if 1 != targetType.NumOut() || "*gowebapi.Response" != targetType.Out(0).String() {

		panic("Invalid target return value (expecting *gowebapi.Response)")
	}

	self.Target = target

	return self
}

func (self *Route) ToMethod(target interface{}, action string) *Route {

	targetType := reflect.TypeOf(target)

	if reflect.Ptr != targetType.Kind() ||
		reflect.Struct != targetType.Elem().Kind() {

		panic("Invalid target type (expecting struct ptr)")
	}

	method, methodExists := targetType.MethodByName(action)

	if !methodExists {

		panic("Invalid target method (method doesn't exist)")
	}

	if 1 != method.Type.NumOut() ||
		"*gowebapi.Response" != method.Type.Out(0).String() {

		panic("Invalid target return value (expecting *gowebapi.Response)")
	}

	self.Target = target
	self.Action = action

	return self
}

func (self *Route) WithFilter(filter filterFunc) *Route {

	self.Filter.Add(filter)

	return self
}
