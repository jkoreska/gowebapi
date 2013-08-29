package gowebapi

import (
	"errors"
	"regexp"
)

type Router interface {
	Route(*Request) (*Route, error)
	AddRoute(string) *Route
	AddRestRoutes(string, interface{})
}

type DefaultRouter struct {
	routes []*Route
}

func (self *DefaultRouter) Route(request *Request) (*Route, error) {

	route := self.getRoute(request.Http.Method, request.Http.URL.Path)

	if nil == route {
		return nil, errors.New("No matching routes")
	}

	params := route.Path.FindStringSubmatch(request.Http.URL.Path)
	for i, param := range params {
		if i > 0 && "" != param {
			route.Params[route.Path.SubexpNames()[i]] = param
		}
	}

	return route, nil
}

func (self *DefaultRouter) AddRoute(path string) *Route {

	params := regexp.MustCompile("{([^}]*)}")
	matches := params.FindAllStringSubmatch(path, -1)

	for _, match := range matches {
		parse := regexp.MustCompile("{" + match[1] + "}")
		path = parse.ReplaceAllString(path, "(?P<"+match[1]+">[^/]*)")
	}

	return self.addRoute(
		&Route{Path: regexp.MustCompile(path), Params: map[string]string{}},
	)
}

func (self *DefaultRouter) AddRestRoutes(path string, controller interface{}) {

	self.AddRoute(path).
		ForMethod("get").
		ToMethod(controller, "Get")
	self.AddRoute(path).
		ForMethod("post").
		ToMethod(controller, "Post")
	self.AddRoute(path).
		ForMethod("put").
		ToMethod(controller, "Put")
	self.AddRoute(path).
		ForMethod("delete").
		ToMethod(controller, "Delete")
}

func (self *DefaultRouter) getRoute(method string, path string) *Route {

	for _, route := range self.routes {

		if route.Path.MatchString(path) {

			if nil == route.Method || route.Method.MatchString(method) {

				return route
			}
		}
	}

	return nil
}

func (self *DefaultRouter) addRoute(route *Route) *Route {

	if nil == self.routes {
		self.routes = make([]*Route, 0)
	}

	self.routes = append(self.routes, route)

	return route
}
