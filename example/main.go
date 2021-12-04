package main

import (
	"fmt"
	"github.com/chenhuazhong/kubehook"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	h := kubehook.Default()
	// validating webhook
	h.Validating("/pod/validate", kubehook.ValidateFun{
		ValidateUpdate: func(obj, old_obj runtime.Object) kubehook.RST {
			pod := obj.(*v1.Pod)
			for _, container := range pod.Spec.Containers {
				if container.ImagePullPolicy != "IfNotPresent" {
					return kubehook.RST{
						Code:    400,
						Result:  false,
						Message: "pass",
					}
				}

			}
			return kubehook.RST{
				Code:    200,
				Result:  true,
				Message: "ok",
			}

		},
		ValidateCreate: func(obj runtime.Object) kubehook.RST {
			pod := obj.(*v1.Pod)
			for _, container := range pod.Spec.Containers {
				if container.ImagePullPolicy != "IfNotPresent" {
					return kubehook.RST{
						Code:    400,
						Result:  false,
						Message: "pass",
					}
				}

			}
			return kubehook.RST{
				Code:    200,
				Result:  true,
				Message: "ok",
			}
		},
		ValidateDelete: func(obj runtime.Object) kubehook.RST {
			return kubehook.RST{
				Code:    200,
				Result:  true,
				Message: "ok",
			}
		},
	})
	// mutating webhook
	h.Mutating("/pod/mutating", func(obj runtime.Object) runtime.Object {
		return obj
	})
	// readness
	h.Route("/health", func(ctx *kubehook.Ctx) {
		ctx.Response(200, []byte("ok"))
	})
	h.Run(fmt.Sprintf("%s:%s", "0.0.0.0", "8080"), "/cert.pem", "/key.pem")
}
