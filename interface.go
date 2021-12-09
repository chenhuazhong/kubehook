package kubehook

import (
	"k8s.io/apimachinery/pkg/runtime"
)

type MiddleWare interface {
	Process_request(ctx *Ctx)

	Process_response(ctx *Ctx)
}

type ResourceHook interface {
	ValidateCreate(obj runtime.Object) RST
	ValidateUpdate(obj, old_obj runtime.Object) RST
	ValidateDelete(obj runtime.Object) RST
}

type ValidataObject interface {
	runtime.Object
	ValidateCreate() RST
	ValidateUpdate(old_obj runtime.Object) RST
	ValidateDelete(obj runtime.Object) RST
}

type MutatingObject interface {
	runtime.Object
	Mutating()
}
