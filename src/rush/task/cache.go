package task

import (
	"github.com/pkg/errors"
	"time"
	"context"
	"sync"
)
// Job of my cloud is to facilitate communication across network to minimize requests required for checkout
// Ensure human-like behaviour

var ErrMiss = errors.New("MISS")

type CloudCacheEntry struct {
	Key string
	Value []byte
	CreatedAt time.Time
	MaxAge int64
}

type CloudCache struct {
	mut sync.Mutex
	entries map[string]*CloudCacheEntry
	getJwt func() string
}

func NewCloudCache(getJwt func() string) *CloudCache {
	return &CloudCache{
		mut: sync.Mutex{},
		entries: map[string]*CloudCacheEntry{},
		getJwt: getJwt,
	}
}

func (c *CloudCache) Get(key string) ([]byte, error) {
	return c.GetContext(context.Background(), key)
}

func (c *CloudCache) GetContext(ctx context.Context, key string) ([]byte, error) {
	c.mut.Lock()
	defer c.mut.Unlock()

	if entry, ok := c.entries[key]; ok {
		return entry.Value, nil
	} else {
		return []byte{}, ErrMiss
	}
}

func (c *CloudCache) CachePush(key string, val []byte) {
	c.mut.Lock()
	defer c.mut.Unlock()
	c.entries[key] = &CloudCacheEntry{
		Key: key,
		Value: val,
		CreatedAt: time.Now(),
	}
}