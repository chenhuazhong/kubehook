# kube-hook
kubernetes admission webhook framework

### install
```shell script
go get -d github.com/chenhuazhong/kubehook
```

### example
```go
package main

func main() {
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
        sidecar := v1.Container{
            Name: "sidecar",
            Image: "nginx",
        }
        pod.Spec.Containers = append(pod.Spec.Containers, sidecar)
        return pod
    })
    // h.Mutating("/guestbook", &GuestBook{}, func(obj runtime.Object) runtime.Object {
	//	return obj
	//})
    h.Route("/health", func(ctx *Ctx) {
        ctx.Response(200, []byte("ok"))
    })
    h.Run(fmt.Sprintf("%s:%s", "0.0.0.0", "8080"), "cert.pem", "key.pem")
}
```