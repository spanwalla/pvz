package metrics

import "github.com/prometheus/client_golang/prometheus"

//go:generate go tool mockgen -source=metrics.go -destination=mocks/mock_metrics.go -package=mocks

type Counter interface {
	Inc()
}

type Counters struct {
	PointsCreated     Counter
	ProductsCreated   Counter
	ReceptionsCreated Counter
}

type PrometheusCounter struct {
	counter prometheus.Counter
}

func NewPrometheusCounter(name, help string) *PrometheusCounter {
	c := &PrometheusCounter{
		counter: prometheus.NewCounter(prometheus.CounterOpts{
			Name: name,
			Help: help,
		}),
	}
	prometheus.MustRegister(c.counter)
	return c
}

func (p *PrometheusCounter) Inc() {
	p.counter.Inc()
}

func New() *Counters {
	return &Counters{
		PointsCreated:     NewPrometheusCounter("points_created_total", "Number of points created"),
		ProductsCreated:   NewPrometheusCounter("products_created_total", "Number of products created"),
		ReceptionsCreated: NewPrometheusCounter("receptions_created_total", "Number of receptions created"),
	}
}
