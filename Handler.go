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
	AddFormatter(string, RequestFormatter, ResponseFormatter)
	ClearFormatters()
}

type DefaultHandler struct {
	router             Router
	binder             Binder
	auther             Auther
	requestFormatters  map[string]RequestFormatter
	responseFormatters map[string]ResponseFormatter
}

func NewDefaultHandler() Handler {

	jsonFormatter := &JsonFormatter{}

	return &DefaultHandler{
		router:             &DefaultRouter{},
		binder:             &DefaultBinder{},
		auther:             &DefaultAuther{},
		requestFormatters:  map[string]RequestFormatter{"application/json": jsonFormatter},
		responseFormatters: map[string]ResponseFormatter{"application/json": jsonFormatter},
	}
}

func (self *DefaultHandler) Router() Router {

	return self.router
}

func (self *DefaultHandler) AddFormatter(mimeType string, requestFormatter RequestFormatter, responseFormatter ResponseFormatter) {

	if nil != requestFormatter {
		self.requestFormatters[mimeType] = requestFormatter
	}
	if nil != responseFormatter {
		self.responseFormatters[mimeType] = responseFormatter
	}
}

func (self *DefaultHandler) ClearFormatters() {

	self.requestFormatters = map[string]RequestFormatter{}
	self.responseFormatters = map[string]ResponseFormatter{}
}

func (self *DefaultHandler) ServeHTTP(responseWriter http.ResponseWriter, httpRequest *http.Request) {

	self.handleResponse(self.handleRequest(httpRequest), responseWriter)
}

func (self *DefaultHandler) handleRequest(httpRequest *http.Request) *Response {

	responseFormat := self.determineResponseFormat(httpRequest.Header)

	if "" == responseFormat {
		return &Response{
			Status: 415,
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
			Data:   routeError,
		}
	}

	request.Route = route

	// read URL query params into request.Data?

	request.Body, _ = ioutil.ReadAll(httpRequest.Body)
	// handle read err

	if len(request.Body) > 0 {

		requestFormat := self.determineRequestFormat(httpRequest.Header)

		if "" == requestFormat {
			return &Response{
				Status: 415,
				Format: responseFormat,
				Data:   "Request format not supported",
			}
		}

		requestFormatter := self.requestFormatters[requestFormat]
		request.Data, _ = requestFormatter.FormatRequest(request.Body)
		// handle format err
	}

	if !self.auther.Authorize(request) {
		return &Response{
			Status: 403,
			Format: responseFormat,
			Data:   "Not authorized",
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

	response.Format = responseFormat

	return response
}

func (self *DefaultHandler) handleResponse(response *Response, responseWriter http.ResponseWriter) {

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
	response.Body, _ = responseFormatter.FormatResponse(response.Data)
	// handle format err

	for header, values := range response.Header {
		for _, value := range values {
			responseWriter.Header().Add(header, value)
		}
	}

	responseWriter.Header().Set("content-type", response.Format)
	responseWriter.WriteHeader(response.Status)
	responseWriter.Write(response.Body)
}

func (self *DefaultHandler) determineRequestFormat(header http.Header) string {

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

func (self *DefaultHandler) determineResponseFormat(header http.Header) string {

	var mimeType string

	if accept := header.Get("accept"); "" != accept && "*/*" != accept {
		accepts := strings.Split(accept, ";")
		for _, accept = range accepts {
			if _, exists := self.responseFormatters[accept]; exists {
				mimeType = accept
			}
		}
	} else {
		for mimeType, _ = range self.responseFormatters {
			break // grab the first one
		}
	}

	return mimeType
}
