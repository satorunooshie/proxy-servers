package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func main() {
	log.SetFlags(log.LUTC | log.Lshortfile)

	addr := flag.String("addr", "127.0.0.1:9999", "proxy address")
	flag.Parse()

	http.HandleFunc("/", proxyHandler)
	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func proxyHandler(w http.ResponseWriter, r *http.Request) {
	// Host must be the absolute path.
	target, err := url.Parse(r.URL.Scheme + "://" + r.URL.Host)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	req, err := httputil.DumpRequest(r, true)
	if err != nil {
		log.Println(err)
		http.Error(w, "Server Error", http.StatusInternalServerError)
		return
	}
	log.Printf("%s\n", req)
	httputil.NewSingleHostReverseProxy(target).ServeHTTP(w, r)
}
