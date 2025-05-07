package rueidisleader

import "github.com/prometheus/client_golang/prometheus"

type MetricsOpts struct {
	IsLeader prometheus.Gauge
	Attempts prometheus.Counter
	Renewals prometheus.Counter
}
