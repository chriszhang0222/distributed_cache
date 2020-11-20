package cache

import (
	"distributed_cache/singleflight"
	"fmt"
	"log"
	"sync"
	pb "distributed_cache/geecachepb"
)

type Group struct{
	name string
	getter Getter
	mainCache cache
	peers PeerPicker
	loader *singleflight.Group
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
		loader: &singleflight.Group{},
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
	return g.load(key)
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

func (g *Group) getFromPeer(peer PeerGetter, key string)(byteView, error){
	//bytes, err := peer.Get(g.name, key)
	req := &pb.Request{
		Group: g.name,
		Key: key,
	}
	res := &pb.Response{}
	err := peer.GetRPC(req, res)

	if err != nil{
		return byteView{}, err
	}
	return byteView{b: res.Value}, nil
}
func (g *Group) load(key string)(byteView, error){
	viewi, err := g.loader.Do(key, func() (interface{}, error) {
		if g.peers != nil{
			if peer, ok := g.peers.PickPeer(key); ok{
				if value, err := g.getFromPeer(peer, key);err == nil{
					return value, nil
				}
				log.Println("[Cache] Failed to get from peer")
			}
		}
		return g.getLocally(key)
	})
	if err == nil{
		return viewi.(byteView), nil
	}
	return byteView{}, err

}


