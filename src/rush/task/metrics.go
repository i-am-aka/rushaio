package task

import "sync"
type TaskMetrics struct {
	Metrics map[string]interface{}
	mut sync.Mutex
}

func (t *TaskMetrics) Incr(key string) {
	t.IncrBy(key, 1)
}

func (t *TaskMetrics) IncrBy(key string, by uint64) {
	t.mut.Lock()
	defer t.mut.Unlock()

	if v, ok := t.Metrics[key]; ok {
		if n, ok := v.(uint64); ok {
			t.Metrics[key] = n + by
		}
	} else {
		t.Metrics[key] = by
	}
}

func (t *TaskMetrics) Get(key string) interface{} {
	t.mut.Lock()
	defer t.mut.Unlock()
	return t.Metrics[key]
}

func (t *TaskMetrics) GetUint64(key string) uint64 {
	t.mut.Lock()
	defer t.mut.Unlock()
	if n, ok := t.Metrics[key].(uint64); ok {
		return n
	} else {
		return 0
	}
}

func (t *TaskMetrics) Set(key string, val uint64) {
	t.mut.Lock()
	defer t.mut.Unlock()
	t.Metrics[key] = val
}

func (t *TaskMetrics) Copy() map[string]interface{} {
	t.mut.Lock()
	defer t.mut.Unlock()
	mc := map[string]interface{}{}
	for k, v := range t.Metrics {
		mc[k] = v
	}
	return mc
}