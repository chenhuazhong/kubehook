package utils

import "net/http"

type Reponse struct {
	http.ResponseWriter
	StatuCode int
}

//func (res *Reponse)SetHeader(key, values string)  {
//	res.Header().Set(key, values)
//}

func (res *Reponse) WriteReponse(data []byte, StatuCode int) (int, error) {
	res.ResponseWriter.WriteHeader(StatuCode)
	res.StatuCode = StatuCode
	return res.ResponseWriter.Write(data)

}
