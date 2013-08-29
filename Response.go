package gowebapi

import (
	"net/http"
)

type Response struct {
	Status int
	http.Header
	Format string
	Body   []byte
	Data   interface{}
}
