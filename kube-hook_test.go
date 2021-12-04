package kubehook

import (
	"fmt"
	v12 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"testing"
)

func TestDefault(t *testing.T) {
	h := Default()
	h.Validating("/validate", ValidateFun{
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
	// 向框架注册自定义资源
	//h.Registry(v13.GroupVersionKind{
	//	Group:   "webapp.my.domain",
	//	Version: "v1",
	//	Kind:    "Guestbook",
	//}, &Guestbook{})
	h.Mutating("/pod-mutating-sudecar?timeout=30s", func(obj runtime.Object) runtime.Object {
		pod := obj.(*v1.Pod)

		pod.Spec.Containers[0].Name = "test"
		return pod
	})
	h.Mutating("/pod-mutating", func(obj runtime.Object) runtime.Object {
		var obj_ runtime.Object
		switch o := obj.(type) {
		case *v12.Deployment:
			fmt.Println("deployment")
			fmt.Println(o.GroupVersionKind())
			obj_ = o
		case *v12.StatefulSet:
			fmt.Println("statefulset")
			fmt.Println(o.GroupVersionKind())
			obj_ = o
		case *v12.DaemonSet:
			fmt.Println("daemonset")
			fmt.Println(o.GroupVersionKind())
			obj_ = o
		}
		return obj_
	})
	h.Route("/health", func(ctx *Ctx) {
		ctx.Response(200, []byte("ok"))
	})
	h.Run(fmt.Sprintf("%s:%s", "0.0.0.0", "8088"), "server-cert.pem", "server-key.pem")
}
