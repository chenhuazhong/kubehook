package main

import (
	"fmt"
	kube_hook "github.com/chenhuazhong/kube-hook"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

func main() {
	h := kube_hook.Default()
	h.Validating("/pod/validate", kube_hook.ValidateFun{
		ValidateUpdate: func(obj, old_obj runtime.Object) hook.RST {
			pod := obj.(*v1.Pod)
			for _, container := range pod.Spec.Containers {
				if container.ImagePullPolicy != "IfNotPresent" {
					return hook.RST{
						Code:    400,
						Result:  false,
						Message: "pass",
					}
				}

			}
			return hook.RST{
				Code:    200,
				Result:  true,
				Message: "ok",
			}

		},
		ValidateCreate: func(obj runtime.Object) hook.RST {
			pod := obj.(*v1.Pod)
			for _, container := range pod.Spec.Containers {
				if container.ImagePullPolicy != "IfNotPresent" {
					return hook.RST{
						Code:    400,
						Result:  false,
						Message: "pass",
					}
				}

			}
			return hook.RST{
				Code:    200,
				Result:  true,
				Message: "ok",
			}
		},
		ValidateDelete: func(obj runtime.Object) hook.RST {
			return hook.RST{
				Code:    200,
				Result:  true,
				Message: "ok",
			}
		},
	})
	h.Run(fmt.Sprintf("%s:%s", "0.0.0.0", "8080"), "/cert.pem", "/key.pem")
}
