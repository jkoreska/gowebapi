package gowebapi

import (
	"log"
)

type filterFunc func(*Request, *Response) (*Response)

type Filter struct {
	filters []filterFunc
}

func (self *Filter) Add(value filterFunc) *Filter {

	self.filters = append(self.filters, value)

	return self
}

func (self *Filter) All() []filterFunc {
	return self.filters
}

func LogFilter(request *Request, response *Response) (*Response) {

	if nil != response {

		log.Printf(
			"%s %s %s %d %d",
			request.Http.RemoteAddr,
			request.Http.Method,
			request.Http.URL.Path,
			response.Status,
			len(response.Body),
			)
	}

	return nil
}

func CorsFilter(request *Request, response *Response) (*Response) {

	if nil == response {
		response = &Response{}
	}
	
	if nil == response.Header {
		response.Header = make(map[string][]string, 0)
	}

	response.Header["Access-Control-Allow-Origin"] = []string{"*"}
	response.Header["Access-Control-Allow-Headers"] = []string{"Accept,Authorization,Origin,Content-type,X-Requested-With"}
	response.Header["Access-Control-Allow-Methods"] = []string{"GET,POST,PUT,DELETE,OPTIONS"}

	if "OPTIONS" == request.Http.Method {
		response.Status = 204
		return response
	}

	return nil
}
