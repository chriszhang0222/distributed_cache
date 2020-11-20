package main

import (
	"distributed_cache/cache"
	"flag"
	"fmt"
	"log"
	"net/http"
	"distributed_cache/cache_http"
)
type Test interface {
	Test(x, y int)int
}

type TestFunc func(x, y int)int

func (f TestFunc)Test(x, y int) int{
	return f(x, y)
}
type Server int

func (h *Server) ServeHTTP(w http.ResponseWriter, r *http.Request){
	log.Println(r.URL.Path)
	w.Write([]byte("Hello"))
}

var db = map[string]string {
	"Tom": "630",
	"Jack": "589",
	"Sam": "567",
}

func createGroup() *cache.Group{
	return cache.NewGroup("scores", 1024, cache.GetterFunc(func(key string) ([]byte, error) {
		log.Println("[SlowDB] search key", key)
		if v, ok := db[key];ok{
			return []byte(v), nil
		}
		return nil, fmt.Errorf("%s not exist", key)
	}))
}

func startCacheServer(addr string, addrs[] string, group *cache.Group){
	peers := cache_http.NewHTTPPool(addr)
	peers.Set(addrs...)
	group.RegisterPeers(peers)
	log.Println("Cache is running at: ", addr)
	log.Fatal(http.ListenAndServe(addr[7:], peers))
}

func startAPIServer(apiAddr string, group *cache.Group){
	http.Handle("/api", http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		key := request.URL.Query().Get("key")
		view, err := group.Get(key)
		if err != nil{
			http.Error(writer, err.Error(), http.StatusInternalServerError)
		}
		writer.Header().Set("Content-Type", "application/octet-stream")
		writer.Write(view.ByteSlice())
	}))
	log.Println("fontend server is running at: ", apiAddr)
	log.Fatal(http.ListenAndServe(apiAddr[7:], nil))
}
func main(){

	var port int
	var api bool
	flag.IntVar(&port, "port", 8003, "Cache server port")
	flag.BoolVar(&api, "api", true, "Start a api server?")
	apiAddr := "http://localhost:9999"

	addrMap := map[int]string{
		8001: "http://localhost:8001",
		8002: "http://localhost:8002",
		8003: "http://localhost:8003",
	}

	var addrs []string
	for _, v := range addrMap{
		addrs = append(addrs, v)
	}
	group := createGroup()
	if api{
		go startAPIServer(apiAddr, group)
	}
	startCacheServer(addrMap[port], []string(addrs), group)

}
