package kubehook

import (
	"encoding/json"
	"github.com/chenhuazhong/kube-hook/utils"
	"gomodules.xyz/jsonpatch/v2"
	v12 "k8s.io/api/admission/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/klog/v2"
	"net/http"
	"strconv"
	"strings"
	"time"
)

func Default(middlewares ...MiddleWare) *Hook {
	s := &Hook{
		Middleware: middlewares,
		HandlerFun: make(map[string]func(ctx *Ctx)),
	}
	return s
}

type Hook struct {
	urlparams  UrlParams
	HandlerFun map[string]func(ctx *Ctx)
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
	h.HandlerFun[url] = f
}

func (h *Hook) Mutating(url string, f Mutatingfun) {
	l := strings.SplitN(url, "?", 2)
	uri := l[0]
	h.HandlerFun[uri] = HandleMutatingFun(f)

}

func (h *Hook) Validating(url string, f ValidateFun) {
	l := strings.SplitN(url, "?", 2)
	uri := l[0]
	h.HandlerFun[uri] = HandleVlidatingFun(f)
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
		}
	}()

	uri_params_list := strings.SplitN(ctx.Request.RequestURI, "?", 2)
	uri := uri_params_list[0]
	if _, ok := h.HandlerFun[uri]; !ok {
		//return 404
		ctx.Response(404, []byte("404 not found"))
		klog.Warningf("%s 404 not found", ctx.Request.RequestURI)
		return nil
	}

	ctx.MiddlewareIndex = 0
	if len(h.Middleware) > 0 {
		for ; ctx.MiddlewareIndex < len(h.Middleware); ctx.MiddlewareIndex++ {
			err = h.NextMiddleware(ctx)
			if err != nil {
				klog.Error(err)
				break
			}
		}
	}
	if err == nil {
		handlerfun := h.HandlerFun[uri]
		ctx.HandlerFunc = handlerfun
		handlerfun(ctx)
	}
	// func(ctx, obj runtime.Object)
	if len(h.Middleware) > 0 {
		for ctx.MiddlewareIndex--; ctx.MiddlewareIndex >= 0; ctx.MiddlewareIndex-- {
			h.Middleware[ctx.MiddlewareIndex].Process_response(ctx)
		}
	}
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

func (h *Hook) Registry(kind metav1.GroupVersionKind, resource runtime.Object) {
	Index.Registry(kind, resource)
}

//

func HandleVlidatingFun(f ValidateFun) func(ctx *Ctx) {
	return func(ctx *Ctx) {
		ser := utils.NewAdmiSsionReviewHeadler(ctx.Request)
		ser.LoadAdmissionReview()
		ctx.Adm_obj = ser.Adm_obj
		resource, err := Index.Get(ctx.Adm_obj.Request.Kind)

		if err != nil {
			ctx.Response(400, []byte(err.Error()))
			return
		}
		err = json.Unmarshal(ctx.Adm_obj.Request.Object.Raw, resource)
		if err != nil {
			ctx.Response(400, []byte(err.Error()))
			return
		}
		ctx.Object = resource
		var ret RST
		if ctx.Adm_obj.Request.Operation == "CREATE" {
			ret = f.ValidateCreate(ctx.Object)
		} else if ctx.Adm_obj.Request.Operation == "UPDATE" {
			ret = f.ValidateUpdate(ctx.Object, ctx.Old_Object)
		} else {
			ret = f.ValidateDelete(ctx.Object)
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

func HandleMutatingFun(f Mutatingfun) func(ctx *Ctx) {
	return func(ctx *Ctx) {
		ser := utils.NewAdmiSsionReviewHeadler(ctx.Request)
		ser.LoadAdmissionReview()
		ctx.Adm_obj = ser.Adm_obj
		resource, err := Index.Get(ctx.Adm_obj.Request.Kind)
		if err != nil {
			ctx.Response(400, []byte(err.Error()))
			return
		}
		err = json.Unmarshal(ctx.Adm_obj.Request.Object.Raw, resource)
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
	}
}