package kubehook

import (
	v12 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
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
	ValidateUpdate func(obj, old_obj runtime.Object) RST
	ValidateDelete func(obj runtime.Object) RST
	ValidateCreate func(obj runtime.Object) RST
}
