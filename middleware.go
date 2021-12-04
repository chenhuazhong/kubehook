package kubehook

import (
	"encoding/json"
	"github.com/chenhuazhong/kube-hook/utils"
	"gomodules.xyz/jsonpatch/v2"
	v12 "k8s.io/api/admission/v1"
	v13 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog/v2"
)

type AdminMiddleWare struct {
}

func (w *AdminMiddleWare) Process_request(ctx *Ctx) {
	ser := utils.NewAdmiSsionReviewHeadler(ctx.Request)
	ser.LoadAdmissionReview()
	ctx.Adm_obj = ser.Adm_obj

}

func (w *AdminMiddleWare) Process_response(ctx *Ctx) {

	switch ctx.HandlerFunc.(type) {
	case Mutatingfun:
		adm_return := v12.AdmissionReview{}
		var PatchTypeJSONPatch v12.PatchType = "JsonPath"
		data, err := json.Marshal(ctx.ChangeObject)
		if err != nil {
			klog.Error(err)
		}
		patch, e := jsonpatch.CreatePatch(ctx.Adm_obj.Request.Object.Raw, data)
		if e != nil {
			klog.Error(e)
		}
		path_byte_data, _ := json.Marshal(patch)
		adm_return.Response = &v12.AdmissionResponse{
			Patch:     path_byte_data,
			PatchType: &PatchTypeJSONPatch,
			UID:       ctx.Adm_obj.Request.UID,
			Allowed:   true,
		}
		data, err = json.Marshal(adm_return)
		if err != nil {
			klog.Error(err)
		}
		ctx.Response(200, data)
	case ValidateFun:
		adm_return := v12.AdmissionReview{}
		c := &v12.AdmissionResponse{
			Allowed: ctx.Validate_result.Result,
			Result: &metav1.Status{
				Code:    ctx.Validate_result.Code,
				Message: ctx.Validate_result.Message,
			},
		}
		adm_return.Response = c
		adm_return_data, err := json.Marshal(adm_return)
		if err != nil {
			klog.Error(err)
		}
		ctx.Response(200, adm_return_data)
	default:

	}
}

type ResourceMiddleWare struct {
}

func (w *ResourceMiddleWare) Process_request(ctx *Ctx) {
	switch ctx.Adm_obj.Request.Resource.Resource {
	case "pods":
		ctx.Object = &v1.Pod{}
		if ctx.Adm_obj.Request.Operation == "UPDATE" {
			ctx.Old_Object = &v1.Pod{}
		}
	case "deployments":
		ctx.Object = &v13.Deployment{}
		if ctx.Adm_obj.Request.Operation == "UPDATE" {
			ctx.Old_Object = &v13.Deployment{}
		}
	case "statefulsets":
		ctx.Object = &v13.StatefulSet{}
		if ctx.Adm_obj.Request.Operation == "UPDATE" {
			ctx.Old_Object = &v13.StatefulSet{}
		}
	case "daemonsets":
		ctx.Object = &v13.DaemonSet{}
		if ctx.Adm_obj.Request.Operation == "UPDATE" {
			ctx.Old_Object = &v13.DaemonSet{}
		}
	default:
		ctx.Object = &v1.Pod{}
		if ctx.Adm_obj.Request.Operation == "UPDATE" {
			ctx.Old_Object = &v13.DaemonSet{}
		}
	}
	_, _, err := utils.Ser.Decode(ctx.Adm_obj.Request.Object.Raw, nil, ctx.Object)
	if err != nil {
		klog.Error(err)
	}
}

func (w *ResourceMiddleWare) Process_response(ctx *Ctx) {
}
