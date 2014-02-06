package gowebapi

import (
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

type Handler interface {
	http.Handler
	Router() Router
	Filter() *Filter
	AddFormatter(string, RequestFormatter, ResponseFormatter)
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

	if nil != requestFormatter {
		self.requestFormatters[mimeType] = requestFormatter
	}
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

	responseFormat := self.determineResponseFormat(httpRequest.Header)

	if "" == responseFormat {
		return &Response{
			Status: 406,
			Data:   "Response format not supported",
		}
	}

	request := &Request{
		Http: httpRequest,
	}

	route, routeError := self.router.Route(request)

	if nil != routeError {
		return &Response{
			Status: 404,
			Format: responseFormat,
			Data:   routeError.Error(),
		}
	}

	request.Route = route

	// read URL query params into request.Data?

	requestBody, readErr := ioutil.ReadAll(httpRequest.Body)

	if nil != readErr {
		return &Response{
			Status: 500,
			Format: responseFormat,
			Data:   readErr.Error(),
		}
	}

	if len(requestBody) > 0 {

		request.Body = requestBody

		requestFormat := self.determineRequestFormat(httpRequest.Header)

		if "" == requestFormat {
			return &Response{
				Status: 415,
				Format: responseFormat,
				Data:   "Request format not supported",
			}
		}

		requestFormatter := self.requestFormatters[requestFormat]
		requestData, formatErr := requestFormatter.FormatRequest(request.Body)

		if nil != formatErr {
			return &Response{
				Status: 500,
				Format: responseFormat,
				Data:   formatErr.Error(),
			}
		}

		request.Data = requestData
	}

	// run filters
	filterResponses := make([]*Response, 0, 0)

	for _, filter := range append(self.filter.All(), route.Filter.All()...) {

		if response, next := filter(request); nil != response {

			if !next {
				return response
			} else if nil != response {
				filterResponses = append(filterResponses, response)
			}
		}
	}

	response, bindError := self.binder.Bind(request)

	if nil != bindError {
		return &Response{
			Status: 500,
			Format: responseFormat,
			Data:   bindError.Error(),
		}
	}

	// add headers from filter responses
	if nil == response.Header {
		response.Header = make(map[string][]string, 0)
	}
	for _, filterResponse := range filterResponses {
		for header, values := range filterResponse.Header {
			for _, value := range values {
				response.Header[header] = append(response.Header[header], value)
			}
		}
	}

	response.Format = responseFormat

	return response
}

func (self *defaultHandler) handleResponse(response *Response, responseWriter http.ResponseWriter) {

	if "" == response.Format {

		for response.Format, _ = range self.responseFormatters {
			break // grab the first one
		}

		if "" == response.Format {
			responseWriter.WriteHeader(500)
			io.WriteString(responseWriter, "No response formatters available")
			return
		}
	}

	responseFormatter := self.responseFormatters[response.Format]
	responseBody, formatErr := responseFormatter.FormatResponse(response.Data)

	if nil != formatErr {
		responseWriter.WriteHeader(500)
		io.WriteString(responseWriter, formatErr.Error())
		return
	}

	response.Body = responseBody

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
		if _, exists := self.requestFormatters[contentType]; exists {
			mimeType = contentType
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
