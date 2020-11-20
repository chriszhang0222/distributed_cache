package main

import (
	"log"
	"net/http"
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
func main(){

}
