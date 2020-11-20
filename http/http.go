package http

import (
	"distributed_cache/cache"
	"distributed_cache/consisenthash"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strings"
	"sync"
)

const defaultBasePath = "/_geecache/"
const defaultReplicas = 50

type HTTPPool struct{
	self string
	basePath string
	mu sync.Mutex
	peers *consisenthash.Map
	httpGetters map[string]*httpGetter
}


type httpGetter struct{
	baseUrl string
}

func NewHTTPPool(self string) *HTTPPool{
	return &HTTPPool{
		self: self,
		basePath: defaultBasePath,
	}
}

func (p *HTTPPool)Set(peers ...string){
	p.mu.Lock()
	defer p.mu.Unlock()
	p.peers = consisenthash.New(defaultReplicas, nil)
	p.httpGetters = make(map[string]*httpGetter, len(peers))
	for _, peer := range peers{
		p.httpGetters[peer] = &httpGetter{
			baseUrl: peer + p.basePath,
		}
	}
}

func (p *HTTPPool)PickPeer(key string)(cache.PeerGetter, bool){
	p.mu.Lock()
	defer p.mu.Unlock()
	if peer := p.peers.Get(key); peer != "" && peer != p.self {
		p.Log("Pick peer %s", peer)
		return p.httpGetters[peer], true
	}
	return nil, false
}

func (p *HTTPPool) Log(format string, v ...interface{}){
	log.Printf(format, v...)
}

func (p *HTTPPool) ServeHTTP(w http.ResponseWriter, r *http.Request){
	if !strings.HasPrefix(r.URL.Path, p.basePath){
		panic("Http serving unexpected path: " + r.URL.Path)
	}
	p.Log("%s, %s", r.Method, r.URL.Path)
	path := r.URL.Path
	parts := strings.SplitN(path[len(p.basePath):], "/", 2)
	if len(parts) != 2{
		http.Error(w, "bad request", http.StatusBadRequest)
		return
	}

	groupName := parts[0]
	key := parts[1]

	group := cache.GetGroup(groupName)
	if group == nil{
		http.Error(w, "no such group: "+groupName, http.StatusNotFound)
		return
	}
	view, err := group.Get(key)
	if err != nil{
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/octet-stream")
	w.Write(view.ByteSlice())
}

func (h *httpGetter)Get(group string, key string)([]byte, error){
	u := fmt.Sprintf("%v/%v/%v",
		h.baseUrl, url.QueryEscape(group), url.QueryEscape(key))
	res, err := http.Get(u)
	if err != nil{
		return nil, err
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK{
		return nil, fmt.Errorf("server returned: %v", res.Status)
	}
	bytes, err := ioutil.ReadAll(res.Body)
	if err != nil{
		return nil, fmt.Errorf("reading response body: %v", err)
	}
	return bytes, nil
}

var _ cache.PeerGetter = (*httpGetter)(nil)


