package main

import (
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	log.SetFlags(log.LUTC | log.Lshortfile)

	fromAddr := flag.String("from", "127.0.0.1:9090", "proxy's listening address")
	toAddr := flag.String("to", "127.0.0.1:7000", "the first address this proxy will forward to")
	flag.Parse()

	s, err := basicReverseProxy(*fromAddr, *toAddr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Starting proxy server on", s.Addr)
	if err := http.ListenAndServe(s.Addr, s.Handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func basicReverseProxy(from, to string) (*http.Server, error) {
	toURL, err := parseToURL(to)
	if err != nil {
		return nil, err
	}
	return &http.Server{Addr: from, Handler: httputil.NewSingleHostReverseProxy(toURL)}, nil
}

func parseToURL(s string) (*url.URL, error) {
	if !strings.HasPrefix(s, "http") {
		s = "http://" + s
	}
	URL, err := url.Parse(s)
	if err != nil {
		return nil, err
	}
	return URL, nil
}
