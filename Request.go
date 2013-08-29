package gowebapi

import (
	"net/http"
)

type Request struct {
	Http *http.Request
	*Route
	Body     []byte
	Data     map[string]interface{}
	UserData string
}
