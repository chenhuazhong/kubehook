package kubehook

import "net/http"

type Reponse struct {
	http.ResponseWriter
	StatuCode int
	data      []byte
}

//func (res *Reponse)SetHeader(key, values string)  {
//	res.Header().Set(key, values)
//}

func (res *Reponse) WriteReponse(data []byte, StatuCode int) (int, error) {
	res.ResponseWriter.WriteHeader(StatuCode)
	res.StatuCode = StatuCode
	res.data = data
	return res.ResponseWriter.Write(data)
}
