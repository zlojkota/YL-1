package memorystate

import (
	"github.com/zlojkota/YL-1/internal/collector"
	"github.com/zlojkota/YL-1/internal/hashhelper"
	"sync"
)

type MemoryState struct {
	metricMap    map[string]*collector.Metrics
	metricMapMux sync.Mutex
	Hasher       *hashhelper.Hasher
}

func (h *MemoryState) InitHasher(hashKey string) {
	h.Hasher = &hashhelper.Hasher{Key: hashKey}
}

func (h *MemoryState) GetHaser() *hashhelper.Hasher {
	return h.Hasher
}

func (h *MemoryState) MetricMapMuxLock() {
	h.metricMapMux.Lock()
}

func (h *MemoryState) MetricMapMuxUnlock() {
	h.metricMapMux.Unlock()
}

func (h *MemoryState) MetricMap() map[string]*collector.Metrics {
	return h.metricMap
}

func (h *MemoryState) SetMetricMap(metricMap map[string]*collector.Metrics) {
	h.metricMapMux.Lock()
	h.metricMap = metricMap
	h.metricMapMux.Unlock()
}

func (h *MemoryState) MetricMapItem(item string) (*collector.Metrics, bool) {
	res, ok := h.metricMap[item]
	return res, ok
}

func (h *MemoryState) SetMetricMapItem(metricMap *collector.Metrics) {
	h.metricMapMux.Lock()
	h.metricMap[metricMap.ID] = metricMap
	h.metricMapMux.Unlock()
}
