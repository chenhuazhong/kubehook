package kubehook

import (
	"encoding/json"
	"fmt"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http/httptest"
	"testing"
)

func TestDefault(t *testing.T) {
	h := Default()
	ts := httptest.NewServer(h)
	defer ts.Close()
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
}

type PodCopy v1.Pod

func (p *PodCopy) t() {

}
func (tt *PodCopy) Validate() {
	fmt.Println("123")
}

func TestPodCopy(t *testing.T) {
	p := PodCopy{
		Spec: v1.PodSpec{
			Containers: []v1.Container{{
				Name: "test",
			},
			},
		},
	}
	fmt.Println(p.GetName())
	data, er := json.Marshal(&p)
	if er != nil {
		fmt.Println(er)
	} else {
		fmt.Println(string(data))
	}
}
