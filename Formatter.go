package gowebapi

import (
	"encoding/json"
)

type Formatter interface {
	MimeType() string
}

type RequestFormatter interface {
	Formatter
	FormatRequest([]byte) (map[string]interface{}, error)
}

type ResponseFormatter interface {
	Formatter
	FormatResponse(interface{}) ([]byte, error)
}

type JsonFormatter struct{}

func (self *JsonFormatter) MimeType() string {

	return "application/json"
}

func (self *JsonFormatter) FormatRequest(body []byte) (map[string]interface{}, error) {

	var obj map[string]interface{}

	if err := json.Unmarshal(body, &obj); nil != err {
		return nil, err
	}

	return obj, nil
}

func (self *JsonFormatter) FormatResponse(obj interface{}) ([]byte, error) {

	body, err := json.Marshal(obj)

	if nil != err {
		return nil, err
	}

	return body, nil
}
