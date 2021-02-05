package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var Counter = promauto.NewCounterVec(prometheus.CounterOpts{Name: "posts"}, []string{"ticker"})

var UpvotesPerHour = promauto.NewHistogram(prometheus.HistogramOpts{Name: "upvoteratio", Buckets: []float64{.005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10, 20, 40, 80, 160, 320, 640}})
