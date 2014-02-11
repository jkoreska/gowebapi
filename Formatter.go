package gowebapi

import (
	"fmt"
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

type TextFormatter struct{}

func (self *TextFormatter) MimeType() string {

	return "text/plain"
}

func (self *TextFormatter) FormatRequest(body []byte) (map[string]interface{}, error) {

	var obj map[string]interface{}

	obj["body"] = string(body)

	return obj, nil
}

func (self *TextFormatter) FormatResponse(obj interface{}) ([]byte, error) {

	switch obj.(type) {
		default:
			return nil, fmt.Errorf("Invalid response.Data type %T", obj)
		case string:
			return []byte(obj.(string)), nil
		case []byte:
			return obj.([]byte), nil
		case nil:
			return []byte(""), nil
	}
}
