package gowebapi

import (
	"fmt"
	"io/ioutil"
	"encoding/json"
)

type Formatter interface {
	MimeType() string
}

type RequestFormatter interface {
	Formatter
	FormatRequest(*Request) error
}

type ResponseFormatter interface {
	Formatter
	FormatResponse(*Response) error
}

type JsonFormatter struct{}

func (self *JsonFormatter) MimeType() string {

	return "application/json"
}

func (self *JsonFormatter) FormatRequest(request *Request) error {

	requestBody, readErr := ioutil.ReadAll(request.Http.Body)

	if nil != readErr {
		return readErr
	}

	request.Body = requestBody

	unmarshalErr := json.Unmarshal(request.Body, &request.Data);

	if nil != unmarshalErr {
		return unmarshalErr
	}

	return nil
}

func (self *JsonFormatter) FormatResponse(response *Response) error {

	body, marshalErr := json.Marshal(response.Data)

	if nil != marshalErr {
		return marshalErr
	}

	response.Body = body

	return nil
}

type TextFormatter struct{}

func (self *TextFormatter) MimeType() string {

	return "text/plain"
}

func (self *TextFormatter) FormatRequest(request *Request) error {

	return nil
}

func (self *TextFormatter) FormatResponse(response *Response) error {

	switch response.Data.(type) {
		default:
			return fmt.Errorf("Invalid response.Data type %T", response.Data)
		case string:
			response.Body = []byte(response.Data.(string))
		case []byte:
			response.Body = response.Data.([]byte)
		case nil:
			response.Body = *new([]byte)
	}

	return nil
}

type NullFormatter struct{}

func (self *NullFormatter) MimeType() string {

	return "*/*"
}

func (self *NullFormatter) FormatRequest(request *Request) error {

	return nil
}

func (self *NullFormatter) FormatResponse(response *Response) error {

	return nil
}
