package gowebapi

import (
	"net/http"
	"strings"
)

type Handler interface {
	http.Handler
	Router() Router
	Filter() *Filter
	AddFormatter(string, RequestFormatter, ResponseFormatter)
	AddRequestFormatter(string, RequestFormatter)
	AddResponseFormatter(string, ResponseFormatter)
	ClearFormatters()
}

type defaultHandler struct {
	router             Router
	binder             Binder
	filter             *Filter
	requestFormatters  map[string]RequestFormatter
	responseFormatters map[string]ResponseFormatter
}

func NewDefaultHandler() Handler {

	jsonFormatter := &JsonFormatter{}

	return &defaultHandler{
		router:             &DefaultRouter{},
		binder:             &DefaultBinder{},
		filter:             &Filter{make([]filterFunc, 0, 0)},
		requestFormatters:  map[string]RequestFormatter{"application/json": jsonFormatter},
		responseFormatters: map[string]ResponseFormatter{"application/json": jsonFormatter},
	}
}

func (self *defaultHandler) Router() Router {

	return self.router
}

func (self *defaultHandler) Filter() *Filter {

	return self.filter
}

func (self *defaultHandler) AddFormatter(mimeType string, requestFormatter RequestFormatter, responseFormatter ResponseFormatter) {

	self.AddRequestFormatter(mimeType, requestFormatter)
	self.AddResponseFormatter(mimeType, responseFormatter)
}

func (self *defaultHandler) AddRequestFormatter(mimeType string, requestFormatter RequestFormatter) {

	if nil != requestFormatter {
		self.requestFormatters[mimeType] = requestFormatter
	}
}

func (self *defaultHandler) AddResponseFormatter(mimeType string, responseFormatter ResponseFormatter) {

	if nil != responseFormatter {
		self.responseFormatters[mimeType] = responseFormatter
	}
}

func (self *defaultHandler) ClearFormatters() {

	self.requestFormatters = map[string]RequestFormatter{}
	self.responseFormatters = map[string]ResponseFormatter{}
}

func (self *defaultHandler) ServeHTTP(responseWriter http.ResponseWriter, httpRequest *http.Request) {

	self.handleResponse(self.handleRequest(httpRequest), responseWriter)
}

func (self *defaultHandler) handleRequest(httpRequest *http.Request) *Response {

	request := &Request{
		Http: httpRequest,
	}

	responseFormat := self.determineResponseFormat(httpRequest.Header)

	if "" == responseFormat {
		return &Response{
			Status: 406,
			Data:   "Response format not supported",
		}
	}

	if httpRequest.ContentLength > 0 {

		requestFormat := self.determineRequestFormat(httpRequest.Header)

		if "" == requestFormat {
			return &Response{
				Status: 415,
				Format: responseFormat,
				Data:   "Request format not supported",
			}
		}

		requestFormatter := self.requestFormatters[requestFormat]
		formatErr := requestFormatter.FormatRequest(request)

		if nil != formatErr {
			return &Response{
				Status: 500,
				Format: responseFormat,
				Data:   formatErr.Error(),
			}
		}
	}

	route, routeError := self.router.Route(request)
	request.Route = route

	if nil != routeError {
		return &Response{
			Status: 404,
			Format: responseFormat,
			Data:   routeError.Error(),
		}
	}

	for _, filter := range append(self.filter.All(), route.Filter.All()...) {
		if responseOverride := filter(request, nil); nil != responseOverride {
			return responseOverride
		}
	}

	response, bindError := self.binder.Bind(request)

	if nil != bindError {
		return &Response{
			Status: 400,
			Format: responseFormat,
			Data:   bindError.Error(),
		}
	}

	if "" == response.Format {
		response.Format = responseFormat
	}

	responseFormatter := self.responseFormatters[response.Format]

	if nil == responseFormatter {
		return &Response{
			Status: 500,
			Data: "No response formatters available for specified response.Format",
		}
	}

	formatErr := responseFormatter.FormatResponse(response)

	if nil != formatErr {
		return &Response{
			Status: 500,
			Data: formatErr.Error(),
		}
	}

	for _, filter := range append(self.filter.All(), route.Filter.All()...) {
		if responseOverride := filter(request, response); nil != responseOverride {
			return responseOverride
		}
	}

	return response
}

func (self *defaultHandler) handleResponse(response *Response, responseWriter http.ResponseWriter) {

	for header, values := range response.Header {
		for _, value := range values {
			responseWriter.Header().Add(header, value)
		}
	}

	responseWriter.Header().Set("content-type", response.Format)
	responseWriter.WriteHeader(response.Status)
	responseWriter.Write(response.Body)
}

func (self *defaultHandler) determineRequestFormat(header http.Header) string {

	var mimeType string

	if contentType := header.Get("content-type"); "" != contentType {
		parts := strings.Split(contentType, ";")
		if _, exists := self.requestFormatters[parts[0]]; exists {
			mimeType = parts[0]
		}
	} else {
		for mimeType, _ = range self.requestFormatters {
			break // grab the first one
		}
	}

	return mimeType
}

func (self *defaultHandler) determineResponseFormat(header http.Header) string {

	var mimeType string

	if accept := header.Get("accept"); "" != accept {

		accepts := strings.Split(accept, ",")

		for _, accept = range accepts {
			accept = strings.Trim(accept, " ")
			parts := strings.Split(accept, ";")

			if _, exists := self.responseFormatters[parts[0]]; exists {
				mimeType = parts[0]
				break
			}
			if "*/*" == parts[0] {
				for mimeType, _ = range self.responseFormatters {
					break // grab the first one
				}
			}
		}
	} else {
		for mimeType, _ = range self.responseFormatters {
			break // grab the first one
		}
	}

	return mimeType
}
