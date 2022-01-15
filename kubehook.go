package kubehook

import (
	"encoding/json"
	"gopkg.in/yaml.v2"
	v1 "k8s.io/api/admissionregistration/v1"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"gomodules.xyz/jsonpatch/v2"
	v12 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
)

func Default(middlewares ...MiddleWare) *Hook {
	s := &Hook{
		Middleware: middlewares,
		handlerFun: make(map[string]webHook),
	}
	return s
}

type Hook struct {
	urlparams  UrlParams
	handlerFun map[string]webHook
	Middleware []MiddleWare
}

func (h *Hook) Run(addr, certFile, keyFile string) {
	klog.Infof("start servetls")
	klog.Infof("cert.pem path: %s", certFile)
	klog.Infof("key.pem path: %s", keyFile)
	klog.Infof("hook server listening at: %s", addr)

	err := http.ListenAndServeTLS(addr, certFile, keyFile, h)
	if err != nil {
		klog.Error(err)
	}
}

func (h *Hook) Route(url string, f func(ctx *Ctx)) {
	h.handlerFun[url] = &WebHook{HandlerFun: f}
}

func (h *Hook) Mutating(url string, resource runtime.Object, f Mutatingfun) {
	l := strings.SplitN(url, "?", 2)
	uri := l[0]
	h.handlerFun[uri] = &MutatingWebhook{
		Object:      resource,
		Mutatingfun: f,
	}
}

func (h *Hook) Validating(url string, resource runtime.Object, f ValidateFun) {
	l := strings.SplitN(url, "?", 2)
	uri := l[0]
	h.handlerFun[uri] = &ValidateWebhook{
		Object:         resource,
		ValidateCreate: f.ValidateCreate,
		ValidateUpdate: f.ValidateUpdate,
		ValidateDelete: f.ValidateDelete,
	}
}

func (h *Hook) Mutatingv1(url string, resource MutatingObject) {
	l := strings.SplitN(url, "?", 2)
	uri := l[0]
	h.Route(uri, HandleMutatingFunv1(resource))
}

func (h *Hook) Validatingv1(url string, resource ValidataObject) {
	l := strings.SplitN(url, "?", 2)
	uri := l[0]
	h.Route(uri, HandleVlidatingFunv1(resource))
}

func (h *Hook) Query() UrlParams {
	return h.urlparams
}

func (h *Hook) NextMiddleware(ctx *Ctx) (err error) {
	defer func() {
		if er := recover(); er != nil {
			err = er.(error)
		}
	}()
	h.Middleware[ctx.MiddlewareIndex].Process_request(ctx)
	return
}

func (h *Hook) HandleFun(ctx *Ctx) (err error) {
	defer func() {
		if e := recover(); e != nil {

			err = e.(error)
			klog.Error(err)
		}
	}()

	uri_params_list := strings.SplitN(ctx.Request.RequestURI, "?", 2)
	uri := uri_params_list[0]
	if _, ok := h.handlerFun[uri]; !ok {
		//return 404
		ctx.Response(404, []byte("404 not found"))
		klog.Warningf("%s 404 not found", ctx.Request.RequestURI)
		return nil
	}
	h.handlerFun[uri].do(ctx)
	return nil
}

func (h *Hook) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	timeout := r.URL.Query().Get("timeout")
	timeout_int, err := strconv.ParseInt(timeout, 10, 0)
	timeout_int = timeout_int

	if err != nil {
		timeout_int = 5
	}
	// todo response write clock
	time_out_ctx := NewContext(time.Duration(timeout_int)*time.Second, w, r)
	done := make(chan int, 1)
	go func() {
		err := h.HandleFun(time_out_ctx)
		if err != nil {
			klog.Error(err)
			time_out_ctx.Response(500, []byte("Internal server error"))
			time_out_ctx.send()
		}
		close(done)
	}()
	select {
	case <-time_out_ctx.Done():
		time_out_ctx.Response(400, []byte("{'message': 'time out'}"))
	case <-done:
		time_out_ctx.send()
	}
	klog.Infof("%s %v %s  \n", time_out_ctx.Request.Method, time_out_ctx.response.StatuCode, time_out_ctx.Request.RequestURI)

	//http.TimeoutHandler()

	// 1、中间件
	// 1.1 处理认证
	// 1.1 admsion mutating序列化
	// 1.2 admsion validatine序列化
	// 1.3 处理各个k8s内建资源
	// 1.4 拓展 自定义资源 序列化
	// 1.5 自定义 中间件
	//
	// 2、视图函数
}

func (h *Hook) Buildconfiguration(Service, Namespce string, Port int32) ([]byte, []byte, error) {
	Ignore := v1.Ignore
	AllScopes := v1.AllScopes
	mutatingWebhookList := []v1.MutatingWebhook{}
	validatingWebhookList := []v1.ValidatingWebhook{}
	for uri, webhook := range h.handlerFun {
		switch webhook.(type) {
		case *MutatingWebhook:
			mutatingWebhookList = append(mutatingWebhookList, v1.MutatingWebhook{
				Name: uri,
				ClientConfig: v1.WebhookClientConfig{
					CABundle: []byte("cacert"),
					Service: &v1.ServiceReference{
						Path:      &uri,
						Name:      Service,
						Namespace: Namespce,
						Port:      &Port,
					},
				},
				FailurePolicy: &Ignore,
				Rules: []v1.RuleWithOperations{
					{
						Operations: []v1.OperationType{v1.Create, v1.Update},
						Rule: v1.Rule{
							APIGroups:   []string{webhook.GetObject().GetObjectKind().GroupVersionKind().Group},
							APIVersions: []string{webhook.GetObject().GetObjectKind().GroupVersionKind().Version},
							Resources:   []string{webhook.GetObject().GetObjectKind().GroupVersionKind().Kind + "s"},
							Scope:       &AllScopes,
						},
					},
				},
				AdmissionReviewVersions: []string{"v1"},
			},
			)
		case *ValidateWebhook:
			validatingWebhookList = append(validatingWebhookList, v1.ValidatingWebhook{
				Name: uri,
				ClientConfig: v1.WebhookClientConfig{
					CABundle: []byte("cacert"),
					Service: &v1.ServiceReference{
						Path:      &uri,
						Name:      Service,
						Namespace: Namespce,
						Port:      &Port,
					},
				},
				FailurePolicy: &Ignore,
				Rules: []v1.RuleWithOperations{
					{
						Operations: []v1.OperationType{v1.Create, v1.Update},
						Rule: v1.Rule{
							APIGroups:   []string{webhook.GetObject().GetObjectKind().GroupVersionKind().Group},
							APIVersions: []string{webhook.GetObject().GetObjectKind().GroupVersionKind().Version},
							Resources:   []string{webhook.GetObject().GetObjectKind().GroupVersionKind().Kind + "s"},
							Scope:       &AllScopes,
						},
					},
				},
				AdmissionReviewVersions: []string{"v1"},
			},
			)
		}
	}
	mutatingConfig := v1.MutatingWebhookConfiguration{
		metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1",
			Kind:       "MutatingWebhookConfiguration",
		},
		metav1.ObjectMeta{
			Name: "kubehook",
		},
		mutatingWebhookList,
	}
	validatingConfig := v1.ValidatingWebhookConfiguration{
		metav1.TypeMeta{
			APIVersion: "admissionregistration.k8s.io/v1",
			Kind:       "ValidatingWebhookConfiguration",
		},
		metav1.ObjectMeta{
			Name: "kubehook",
		},
		validatingWebhookList,
	}

	muData, err := json.Marshal(mutatingConfig)
	if err != nil {
		return nil, nil, err
	}
	vaData, err := json.Marshal(validatingConfig)
	if err != nil {
		return nil, nil, err
	}
	return muData, vaData, nil
}

func (h *Hook) LoadMutatingWebhookConfiguration(Service, Namespce string, Port int32) {
	muData, vaData, err := h.Buildconfiguration(Service, Namespce, Port)
	webhookconfiglist := []map[string]interface{}{}
	muwebhookconfig := make(map[string]interface{})
	vawebhookconfig := make(map[string]interface{})
	_ = json.Unmarshal(muData, &muwebhookconfig)
	_ = json.Unmarshal(vaData, &vawebhookconfig)
	webhookconfiglist = append(webhookconfiglist, muwebhookconfig, vawebhookconfig)
	Data, err := yaml.Marshal(webhookconfiglist)
	if err != nil {
		klog.Error(err)
	} else {
		f, err := os.OpenFile("./webhook-config.yaml", os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0644)
		if err != nil {
			klog.Error(err)
		} else {
			_, err := f.Write(Data)
			klog.Error(err)
		}
	}
}

func HandleVlidatingFunv2(resource runtime.Object, f ValidateFun) func(ctx *Ctx) {
	return func(ctx *Ctx) {
		adm_obj := v12.AdmissionReview{}
		err := json.Unmarshal(ctx.Request.Data(), &adm_obj)
		if err != nil {
			klog.Error(err)
			ctx.Response(400, []byte(err.Error()))
			return
		}
		err = json.Unmarshal(adm_obj.Request.Object.Raw, resource)
		if err != nil {
			klog.Error(err)
			ctx.Response(400, []byte(err.Error()))
			return
		}
		ctx.Object = resource
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
			// todo return  false
			ctx.Response(400, []byte(err.Error()))
		} else {
			ctx.Response(200, adm_return_data)
		}
	}
}

func HandleMutatingFunv2(resource runtime.Object, f Mutatingfun) func(ctx *Ctx) {
	return func(ctx *Ctx) {
		adm_obj := v12.AdmissionReview{}
		err := json.Unmarshal(ctx.Request.Data(), &adm_obj)
		if err != nil {
			klog.Error(err)
			ctx.Response(400, []byte(err.Error()))
			return
		}
		err = json.Unmarshal(adm_obj.Request.Object.Raw, resource)
		if err != nil {
			ctx.Response(400, []byte(err.Error()))
			return
		}
		ctx.Object = resource

		obj := f(ctx.Object)
		ctx.ChangeObject = obj
		adm_return := v12.AdmissionReview{}
		var PatchTypeJSONPatch v12.PatchType = "JsonPath"
		data, err := json.Marshal(ctx.ChangeObject)
		if err != nil {
			klog.Error(err)
		}
		patch, e := jsonpatch.CreatePatch(adm_obj.Request.Object.Raw, data)
		if e != nil {
			klog.Error(e)
		}
		path_byte_data, _ := json.Marshal(patch)
		adm_return.Response = &v12.AdmissionResponse{
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
}

func HandleMutatingFunv1(obj MutatingObject) func(ctx *Ctx) {
	return func(ctx *Ctx) {
		adm_obj := v12.AdmissionReview{}
		err := json.Unmarshal(ctx.Request.Data(), &adm_obj)
		if err != nil {
			klog.Error(err)
			ctx.Response(400, []byte(err.Error()))
			return
		}
		err = json.Unmarshal(adm_obj.Request.Object.Raw, obj)
		if err != nil {
			ctx.Response(400, []byte(err.Error()))
			return
		}
		ctx.Object = obj.DeepCopyObject()
		obj.Mutating()
		ctx.ChangeObject = obj
		adm_return := v12.AdmissionReview{}
		var PatchTypeJSONPatch v12.PatchType = "JsonPath"
		data, err := json.Marshal(ctx.ChangeObject)
		if err != nil {
			klog.Error(err)
		}
		patch, e := jsonpatch.CreatePatch(adm_obj.Request.Object.Raw, data)
		if e != nil {
			klog.Error(e)
		}
		path_byte_data, _ := json.Marshal(patch)
		adm_return.Response = &v12.AdmissionResponse{
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
}

func HandleVlidatingFunv1(obj ValidataObject) func(ctx *Ctx) {

	return func(ctx *Ctx) {
		adm_obj := v12.AdmissionReview{}
		err := json.Unmarshal(ctx.Request.Data(), &adm_obj)
		if err != nil {
			klog.Error(err)
			ctx.Response(400, []byte(err.Error()))
			return
		}
		err = json.Unmarshal(adm_obj.Request.Object.Raw, obj)
		if err != nil {
			klog.Error(err)
			ctx.Response(400, []byte(err.Error()))
			return
		}
		ctx.Object = obj
		var ret RST
		if adm_obj.Request.Operation == "CREATE" {
			ret = obj.ValidateCreate()
		} else if adm_obj.Request.Operation == "UPDATE" {
			ret = obj.ValidateUpdate(ctx.Old_Object)
		} else {
			ret = obj.ValidateDelete(ctx.Object)
		}
		ctx.Validate_result = ret
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
			// todo return  false
			ctx.Response(400, []byte(err.Error()))
		} else {
			ctx.Response(200, adm_return_data)
		}
	}
}
