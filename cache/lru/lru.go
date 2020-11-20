package lru

import "container/list"

type Cache struct{
	ll *list.List
	maxBytes int64
	nbytes int64
	cache map[string]*list.Element
	OnEvicted func(key string, value Value)
}

type entry struct{
	key   string
	value Value
}

type Value interface {
	Len() int
}


func New(maxBytes int64, onEvicted func(string, Value)) *Cache {
	return &Cache{
		maxBytes: maxBytes,
		OnEvicted: onEvicted,
	}
}

func (c *Cache)Get(key string)(Value, bool){
	if ele, ok := c.cache[key];ok{
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		return kv.value, true

	}
	return nil, false
}

func (c *Cache)RemoveOldest(){
	var ele *list.Element = c.ll.Back()
	if ele != nil{
		c.ll.Remove(ele)
		kv := ele.Value.(*entry)
		delete(c.cache, kv.key)
		c.nbytes -= int64(len(kv.key)) + int64(kv.value.Len())
		if c.OnEvicted != nil{
			c.OnEvicted(kv.key, kv.value)
		}
	}
}

func (c *Cache)Add(key string, value Value){
	if ele, ok := c.cache[key];ok{
		c.ll.MoveToFront(ele)
		kv := ele.Value.(*entry)
		c.nbytes += int64(value.Len()) - int64(kv.value.Len())
		kv.value = value
	}else{
		var ele = c.ll.PushFront(&entry{key: key, value: value})
		c.cache[key] = ele
		c.nbytes += int64(len(key)) + int64(value.Len())
	}
	for c.maxBytes != 0 && c.maxBytes < c.nbytes{
		c.RemoveOldest()
	}
}

func (c *Cache) Len() int {
	return c.ll.Len()
}
