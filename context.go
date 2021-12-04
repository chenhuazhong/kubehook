package kubehook

import (
	"context"
	v1 "k8s.io/api/admission/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"net/http"
	"time"
)

type Ctx struct {
	//context.WithTimeout
	context.Context
	context.CancelFunc
	deadline   time.Time
	time       *time.Timer
	Object     runtime.Object
	Old_Object runtime.Object
	Raw_Object runtime.Object

	ChangeObject    runtime.Object
	row_obj         struct{}
	Adm_obj         v1.AdmissionReview
	value           map[interface{}]interface{}
	Validate_result RST
	HandlerFunc     interface{}
	MiddlewareIndex int
	Request         *http.Request
	response        *Reponse
	data            []byte
}

func (ctx *Ctx) Deadline() (deadline time.Time, ok bool) {
	return ctx.deadline, true
}

func (ctx *Ctx) Cancal() {
	//close(ctx.ch1)
	ctx.CancelFunc()
}

func (ctx *Ctx) Done() <-chan struct{} {
	return ctx.Context.Done()
}

func (ctx *Ctx) Err() error {
	return nil
}

func (ctx *Ctx) Value(key interface{}) interface{} {
	return ctx.value[key]
}

func (ctx *Ctx) Response(status_code int, body []byte) {
	ctx.response.data = body
	ctx.response.StatuCode = status_code
}

func (ctx *Ctx) send() {
	ctx.response.WriteHeader(ctx.response.StatuCode)
	_, _ = ctx.response.Write(ctx.response.data)
}

func NewContext(time_out time.Duration, response http.ResponseWriter, request *http.Request) *Ctx {
	cancel, cancel_fun := context.WithCancel(context.TODO())

	c := &Ctx{
		Context:    cancel,
		CancelFunc: cancel_fun,
		response:   &Reponse{ResponseWriter: response},
		Request:    request,
	}
	c.time = time.AfterFunc(time_out, func() {
		c.CancelFunc()
	})
	return c
}
