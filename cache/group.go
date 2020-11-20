package cache

import (
	"fmt"
	"log"
	"sync"
)

type Group struct{
	name string
	getter Getter
	mainCache cache
	peers PeerPicker
}

var  (
	mu sync.RWMutex
	groups = make(map[string]*Group)
)

func NewGroup(name string, cacheBytes int64, getter Getter)*Group{
	if getter == nil{
		panic("nil getter")
	}
	mu.Lock()
	defer mu.Unlock()
	g := &Group{
		name: name,
		getter: getter,
		mainCache: cache{cacheBytes: cacheBytes},
	}
	groups[name] = g
	return g
}

func GetGroup(name string) *Group{
	mu.RLock()
	g := groups[name]
	mu.RUnlock()
	return g
}

func (g *Group)Get(key string)(byteView, error){
	if key == ""{
		return byteView{}, fmt.Errorf("key is required")
	}
	if v, ok := g.mainCache.get(key); ok{
		log.Println("[GeeCache] hit")
		return v, nil
	}
	return g.getLocally(key)
}

func (g *Group) getLocally(key string)(byteView, error){
	bytes, err := g.getter.Get(key)
	if err != nil{
		return byteView{}, err
	}
	value := byteView{b: cloneBytes(bytes)}
	g.populateCache(key, value)
	return value, nil
}

func (g *Group) populateCache(key string, value byteView){
	g.mainCache.add(key, value)
}

func (g *Group) RegisterPeers(peers PeerPicker){
	if(g.peers != nil){
		panic("Register peer called more than once")
	}
	g.peers = peers
}


