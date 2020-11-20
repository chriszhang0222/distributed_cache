package cache

import (
	"distributed_cache/cache/lru"
	"sync"
)

type cache struct {
	lru *lru.Cache
	mutex sync.Mutex
	cacheBytes int64
}

func (c *cache) add(key string, value byteView){
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil{
		c.lru = lru.New(c.cacheBytes, nil)
	}
	c.lru.Add(key, value)
}

func (c *cache)get(key string)(value byteView, ok bool){
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if c.lru == nil{
		return
	}
	if v, ok := c.lru.Get(key);ok{
		return v.(byteView), ok
	}
	return
}
