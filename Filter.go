package gowebapi

type filterFunc func(*Request) (*Response, bool)

type Filter struct {
	filters []filterFunc
}

func (self *Filter) Add(value filterFunc) {
	self.filters = append(self.filters, value)
}

func (self *Filter) All() []filterFunc {
	return self.filters
}
