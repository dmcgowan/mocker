package main

import (
	"flag"
	"github.com/gorilla/mux"
	"net/http"
	"strconv"
)

var host string
var port int

func init() {
	flag.StringVar(&host, "host", "0.0.0.0", "host to bind api server")
	flag.IntVar(&port, "port", 8080, "port to bind api server")
}

func main() {
	flag.Parse()
	api := NewMockApi()
	router := mux.NewRouter()
	api.Register(router)
	http.ListenAndServe(host+":"+strconv.Itoa(port), router)
}
