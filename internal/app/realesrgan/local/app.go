package local

import (
	"github.com/prometheus/client_golang/prometheus"
)

type RealesrganLocal struct {
	PromNamespace   string
	UpsizeTimeGauge prometheus.Gauge
}

// NewRealesrganLocal is the constructor for running local upsizing.
func NewRealesrganLocal(promNamespace string) RealesrganLocal {

	var upsizeTime = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Namespace: promNamespace,
			Name:      "upsize_time",
			Help:      "time it tool to upsize the image",
		},
	)
	prometheus.MustRegister(upsizeTime)

	return RealesrganLocal{PromNamespace: promNamespace, UpsizeTimeGauge: upsizeTime}
}
