package repository

import (
	"encoding/json"
	"time"

	"github.com/patrickmn/go-cache"
)

type DistCache struct{ c *cache.Cache }

func NewMemcached(addr string) *DistCache { 
	return &DistCache{c: cache.New(5*time.Minute, 10*time.Minute)}
}

func (d *DistCache) Get(key string, out any) bool {
	val, found := d.c.Get(key)
	if !found {
		return false
	}
	if data, ok := val.([]byte); ok {
		return json.Unmarshal(data, out) == nil
	}
	return false
}
func (d *DistCache) Set(key string, val any, ttl time.Duration) {
	b, _ := json.Marshal(val)
	d.c.Set(key, b, ttl)
}
func (d *DistCache) Delete(key string) { d.c.Delete(key) }
