package common

import (
	"time"

	"github.com/gammazero/deque"
)

type Metric struct {
	Timestamp  time.Time
	Latency    int64
	Throughput int64
}

type MetricGroup struct {
	Avg Metric
}

func NewMetricsManager(size int) MetricsManager {
	return MetricsManager{
		size:    size,
		metrics: deque.New[Metric](),
	}
}

type MetricsManager struct {
	size            int
	metrics         *deque.Deque[Metric]
	totalLatency    int64
	totalThroughput int64
}

func (mm *MetricsManager) prune() {
	for mm.metrics.Len() > mm.size {
		metric := mm.metrics.PopFront()
		mm.totalLatency -= metric.Latency
		mm.totalThroughput -= metric.Throughput
	}
}

func (mm *MetricsManager) Process(m Metric) {
	mm.totalLatency += m.Latency
	mm.totalThroughput += m.Throughput
	mm.metrics.PushBack(m)
	mm.prune()
}

func (mm *MetricsManager) Group() (mg MetricGroup) {
	mm.prune()
	if mm.metrics.Len() > 0 {
		mg.Avg = Metric{
			Latency:    mm.totalLatency / int64(mm.metrics.Len()),
			Throughput: mm.totalThroughput / int64(mm.metrics.Len()),
		}
	}
	return mg
}
