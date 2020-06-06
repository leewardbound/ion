package biz

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	statDispatch = promauto.NewCounterVec(prometheus.CounterOpts{
		Name: "ion_biz_protoo_recv_msg",
		Help: "Ion biz server",
	}, []string{"method"})
)
