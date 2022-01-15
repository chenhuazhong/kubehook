package metrics

import (
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"k8s.io/klog/v2"
	"net/http"
)

func MetricsService(addr string) {
	klog.Error(http.ListenAndServe(addr, promhttp.Handler()))

}
