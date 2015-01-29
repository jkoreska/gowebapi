package gowebapi

type filterFunc func(*Request) (*Response, bool)

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

func CorsFilter(request *Request) (*Response, bool) {

	response := &Response{
		Status: 204,
		Header: map[string][]string{
			"Access-Control-Allow-Origin": []string{"*"},
			"Access-Control-Allow-Headers": []string{"Accept,Authorization,Origin,Content-type,X-Requested-With"},
			"Access-Control-Allow-Methods": []string{"GET,POST,PUT,DELETE,OPTIONS"},
		},
	}

	next := "OPTIONS" != request.Http.Method

	return response, next
}
