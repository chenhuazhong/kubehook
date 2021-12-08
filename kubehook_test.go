package kubehook

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestHook_HandleFun(t *testing.T) {
	hook := Hook{}
	hook.Validating("/validating", &v1.Pod{}, ValidateFun{
		ValidateDelete: func(obj runtime.Object) RST {
			return RST{
				Code: 200,
				Result: true,
				Message: "ok",
			}
		},
		ValidateCreate: func(obj runtime.Object) RST {
			return RST{
				Code: 200,
				Result: true,
				Message: "ok",
			}
		},
		ValidateUpdate: func(obj, old_obj runtime.Object) RST {
			return RST{
				Code: 200,
				Result: true,
				Message: "ok",
			}
		},
	})
	hook.Mutating("/mutating", &v1.Pod{}, func(obj runtime.Object) runtime.Object {
		return obj
	})
	err := hook.HandleFun(&Ctx{
		//Request: http.Request{}
	})
	if err != nil{
		t.Error(err)
	}


}


func TestDefault(t *testing.T) {
	h := Default()
	h.Validating("/validate", &v1.Pod{}, ValidateFun{
		ValidateUpdate: func(obj, old_obj runtime.Object) RST {
			return RST{Result: true}
		},
		ValidateDelete: func(obj runtime.Object) RST {
			return RST{Result: true}
		},
		ValidateCreate: func(obj runtime.Object) RST {
			return RST{Result: true}
		},
	})
	h.Mutating("/pod-mutating-sidecar?timeout=30s", &v1.Pod{}, func(obj runtime.Object) runtime.Object {
		pod := obj.(*v1.Pod)
		pod.Spec.Containers[0].Name = "test"
		return pod
	})
	h.Route("/health", func(ctx *Ctx) {
		ctx.Response(200, []byte("ok"))
	})
	h.Run(fmt.Sprintf("%s:%s", "0.0.0.0", "8088"), "server-cert.pem", "server-key.pem")
}
