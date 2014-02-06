package gowebapi

func NewResponse() Response {
	return Response{0, "", nil, nil, make(map[string][]string, 0)}
}

type Response struct {
	Status int
	Format string
	Body   []byte
	Data   interface{}
	Header map[string][]string
}
