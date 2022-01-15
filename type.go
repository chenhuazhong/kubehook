package kubehook

import (
	"encoding/json"
	"gomodules.xyz/jsonpatch/v2"
	admissv1 "k8s.io/api/admission/v1"
	v12 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

const (
	CREATE = "CREATE"
	UPDATE = "UPDATE"
	DELETE = "DELETE"
)

type UrlParams map[string]string

type index map[v1.GroupVersionKind]runtime.Object

func (i index) Registry(gvk v1.GroupVersionKind, obj runtime.Object) {
	i[gvk] = obj
}

func (i index) Get(gvk v1.GroupVersionKind) (runtime.Object, error) {
	if v, ok := i[gvk]; ok {
		return v, nil
	} else {
		return nil, &TypeError{"sd"}
	}
}

var Index = &index{
	v1.GroupVersionKind{
		Group:   "",
		Version: "v1",
		Kind:    "Pod",
	}: &v12.Pod{},
}

type RST struct {
	Code    int32
	Message string
	Result  bool
}

type Mutatingfun func(obj runtime.Object) runtime.Object

type ValidateFun struct {
	ValidateUpdate func(obj, old_obj runtime.Object) error
	ValidateDelete func(obj runtime.Object) error
	ValidateCreate func(obj runtime.Object) error
}

type WebHook struct {
	HandlerFun func(ctx *Ctx)
}

func (w *WebHook) do(ctx *Ctx) {
	w.HandlerFun(ctx)
}

type ValidateWebhook struct {
	runtime.Object
	ValidateUpdate func(obj, old_obj runtime.Object) error
	ValidateDelete func(obj runtime.Object) error
	ValidateCreate func(obj runtime.Object) error
}

func (f *ValidateWebhook) do(ctx *Ctx) {
	adm_obj := admissv1.AdmissionReview{}
	err := json.Unmarshal(ctx.Request.Data(), &adm_obj)
	if err != nil {
		klog.Error(err)
		ctx.Response(400, []byte(err.Error()))
		return
	}
	err = json.Unmarshal(adm_obj.Request.Object.Raw, f.Object)
	if err != nil {
		klog.Error(err)
		ctx.Response(400, []byte(err.Error()))
		return
	}
	ctx.Object = f.Object
	var validateErr error
	if adm_obj.Request.Operation == CREATE {
		validateErr = f.ValidateCreate(ctx.Object)
	} else if adm_obj.Request.Operation == UPDATE {
		validateErr = f.ValidateUpdate(ctx.Object, ctx.Old_Object)
	} else {
		validateErr = f.ValidateDelete(ctx.Object)
	}
	if validateErr != nil {
		ctx.Validate_result = RST{
			Code:    400,
			Message: validateErr.Error(),
			Result:  false,
		}
	} else {
		ctx.Validate_result = RST{
			Code:    200,
			Message: "",
			Result:  true,
		}
	}
	adm_return := admissv1.AdmissionReview{}
	c := &admissv1.AdmissionResponse{
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
		// todo return  false
		ctx.Response(400, []byte(err.Error()))
	} else {
		ctx.Response(200, adm_return_data)
	}
}

type MutatingWebhook struct {
	runtime.Object
	Mutatingfun func(obj runtime.Object) runtime.Object
}

func (w *MutatingWebhook) do(ctx *Ctx) {
	adm_obj := admissv1.AdmissionReview{}
	err := json.Unmarshal(ctx.Request.Data(), &adm_obj)
	if err != nil {
		klog.Error(err)
		ctx.Response(400, []byte(err.Error()))
		return
	}
	err = json.Unmarshal(adm_obj.Request.Object.Raw, w.Object)
	if err != nil {
		ctx.Response(400, []byte(err.Error()))
		return
	}
	ctx.Object = w.Object

	obj := w.Mutatingfun(ctx.Object)
	ctx.ChangeObject = obj
	adm_return := admissv1.AdmissionReview{}
	var PatchTypeJSONPatch admissv1.PatchType = "JsonPath"
	data, err := json.Marshal(ctx.ChangeObject)
	if err != nil {
		klog.Error(err)
	}
	patch, e := jsonpatch.CreatePatch(adm_obj.Request.Object.Raw, data)
	if e != nil {
		klog.Error(e)
	}
	path_byte_data, _ := json.Marshal(patch)
	adm_return.Response = &admissv1.AdmissionResponse{
		Patch:     path_byte_data,
		PatchType: &PatchTypeJSONPatch,
		UID:       adm_obj.Request.UID,
		Allowed:   true,
	}
	data, err = json.Marshal(adm_return)
	if err != nil {
		klog.Error(err)
	}
	ctx.Response(200, data)
}
