package kubehook

import (
	"io/ioutil"
	"net/http"
)

type Request struct {
	*http.Request
	data []byte
	URL  string
}

//func (res *Reponse)SetHeader(key, values string)  {
//	res.Header().Set(key, values)
//}

func (request *Request) Data() []byte {
	return request.data
}

func NewRequest(r *http.Request) *Request {
	data, err := ioutil.ReadAll(r.Body)
	if err != nil {
		data = make([]byte, 0)
	}
	return &Request{
		Request: r,
		data:    data,
		URL:     r.RequestURI,
	}
}
