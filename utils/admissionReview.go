package utils

import (
	"io/ioutil"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"net/http"
)

//codecs := serializer.NewCodecFactory(runtimeScheme)
type AdmissionReviewHeadler struct {
	codecs        serializer.CodecFactory
	runtimeScheme *runtime.Scheme
	Deserializer  runtime.Decoder
	Request       *http.Request
	Resource      APIResource
	Adm_obj       v1.AdmissionReview
}

func Serializer() runtime.Decoder {
	return serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()
}

var Ser = Serializer()

func NewAdmiSsionReviewHeadler(request *http.Request) AdmissionReviewHeadler {
	return AdmissionReviewHeadler{
		Request:      request,
		Deserializer: Ser,
	}

}

//func (adm *AdmissionReviewHeadler) Data() v1.AdmissionReview{
//	return v1.AdmissionReview{
//		Request: adm.Request
//	}
//}
func (adm *AdmissionReviewHeadler) LoadAdmissionReview() {
	adm_json_data, err := ioutil.ReadAll(adm.Request.Body)
	if err != nil {
		panic(err)
	}
	admission_obj, _, decode_err := adm.Deserializer.Decode(adm_json_data, nil, nil)
	switch admission_obj.(type) {
	case *v1.AdmissionReview:
		admin_v1 := admission_obj.(*v1.AdmissionReview)
		adm.Adm_obj = *admin_v1
	case *v1beta1.AdmissionReview:
		admin_v1 := admission_obj.(*v1.AdmissionReview)
		adm.Adm_obj = *admin_v1
	}
	if decode_err != nil {
		panic(err)
	}
}

func (adm *AdmissionReviewHeadler) Load(resou runtime.Object) {
	_, _, err := adm.Deserializer.Decode(adm.Adm_obj.Request.Object.Raw, nil, resou)
	if err != nil {
		panic(err)
	}
}

//
//var (
//	runtimeScheme = runtime.NewScheme()
//	codecs        = serializer.NewCodecFactory(runtimeScheme)
//	deserializer  = codecs.UniversalDeserializer()
//
//	// (https://github.com/kubernetes/kubernetes/issues/57982)
//	defaulter = runtime.ObjectDefaulter(runtimeScheme)
//)
