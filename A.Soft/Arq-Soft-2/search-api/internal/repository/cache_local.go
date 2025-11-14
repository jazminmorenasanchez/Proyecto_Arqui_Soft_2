package repository

import (
	"time"

	"github.com/karlseguin/ccache/v3"
)

type LocalCache struct{ c *ccache.Cache[any] }

func NewLocalCache(size int) *LocalCache {
	return &LocalCache{c: ccache.New(ccache.Configure[any]().MaxSize(int64(size)))}
}
func (l *LocalCache) Get(key string) interface{} {
	if it := l.c.Get(key); it != nil {
		return it.Value()
	}
	return nil
}
func (l *LocalCache) Set(key string, val interface{}, ttl time.Duration) {
	l.c.Set(key, val, ttl)
}
func (l *LocalCache) Delete(key string) { l.c.Delete(key) }
