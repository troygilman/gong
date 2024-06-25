package common

import (
	"github.com/gammazero/deque"
	"time"
)

type Metric struct {
	Timestamp  time.Time
	Latency    int64
	Throughput int64
}

type MetricGroup struct {
	Avg Metric
}

func NewMetricsManager(duration time.Duration) MetricsManager {
	return MetricsManager{
		duration: duration,
	}
}

type MetricsManager struct {
	duration        time.Duration
	metrics         deque.Deque[Metric]
	totalLatency    int64
	totalThroughput int64
}

func (mm *MetricsManager) prune(now time.Time) {
	cutoff := now.Add(-mm.duration)
	for mm.metrics.Back().Timestamp.Before(cutoff) {
		metric := mm.metrics.PopBack()
		mm.totalLatency -= metric.Latency
		mm.totalThroughput -= metric.Throughput
	}
}

func (mm *MetricsManager) Process(m Metric, now time.Time) {
	mm.totalLatency += m.Latency
	mm.totalThroughput += m.Throughput
	mm.metrics.PushFront(m)
	mm.prune(now)
}

func (mm MetricsManager) Group() (mg MetricGroup) {
	if mm.metrics.Len() > 0 {
		mg.Avg = Metric{
			Latency:    mm.totalLatency / int64(mm.metrics.Len()),
			Throughput: mm.totalThroughput / int64(mm.metrics.Len()),
		}
	}
	return mg
}
