package task

import "sync"

type Counter struct {
	mapping map[string]uint32
	mut sync.Mutex
}

func (c *Counter) Add(key string) {
	c.mut.Lock()
	defer c.mut.Unlock()
}