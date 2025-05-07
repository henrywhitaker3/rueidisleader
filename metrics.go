package rueidisleader

import "github.com/prometheus/client_golang/prometheus"

type MetricsOpts struct {
	Namespace string
	Subsystem string
}

type metrics struct {
	isLeader prometheus.Gauge
	attempts prometheus.Counter
	renewals prometheus.Counter
}

func newMetrics(opts MetricsOpts) *metrics {
	return &metrics{
		isLeader: prometheus.NewGauge(prometheus.GaugeOpts{
			Name:      "rueidisleader_is_leader",
			Help:      "Whether the instance is the leader, 1=leader, 0=not the leader",
			Namespace: opts.Namespace,
			Subsystem: opts.Subsystem,
		}),
		attempts: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "rueidisleader_lease_attemps_total",
			Help:      "The number of attempts to acquire the lease",
			Namespace: opts.Namespace,
			Subsystem: opts.Subsystem,
		}),
		renewals: prometheus.NewCounter(prometheus.CounterOpts{
			Name:      "rueidisleader_lease_renewals_total",
			Help:      "The number of attempts to acquire the lease",
			Namespace: opts.Namespace,
			Subsystem: opts.Subsystem,
		}),
	}
}

func (m *metrics) register(reg prometheus.Registerer) {
	m.isLeader.Set(1)
	reg.Register(m.isLeader)
	reg.Register(m.attempts)
	reg.Register(m.renewals)
}
