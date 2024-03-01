package main

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"

	"golang.org/x/net/http2"
)

var ()

func mainHandler(w http.ResponseWriter, r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))
	w.Header().Add("backend", "yes")
	w.Header().Add("test-header", "original")
	fmt.Fprint(w, "responded with test-header")
}

func noHeader(w http.ResponseWriter, r *http.Request) {
	requestDump, err := httputil.DumpRequest(r, true)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(string(requestDump))
	w.Header().Add("backend", "yes")
	fmt.Fprint(w, "responded without test header")
}

func main() {
	address := ":9091"
	http.HandleFunc("/", mainHandler)
	http.HandleFunc("/no", noHeader)
	server := &http.Server{
		Addr: address,
	}
	http2.ConfigureServer(server, &http2.Server{})
	log.Println("Backend server listen at ", address)
	server.ListenAndServe()
}
