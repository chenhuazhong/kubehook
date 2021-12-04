package kubehook

import (
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

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
