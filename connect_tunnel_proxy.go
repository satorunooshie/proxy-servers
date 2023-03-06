package main

import (
	"flag"
	"io"
	"log"
	"net"
	"net/http"
)

// https_proxy=localhost:9999 curl -v https://example.org
// https_proxy=localhost:9999 go run http_get_explicit_proxy.go --target https://example.org
func main() {
	log.SetFlags(log.LUTC | log.Lshortfile)

	addr := flag.String("addr", "127.0.0.1:9999", "proxy address")
	flag.Parse()

	s := new(forwardProxy)

	log.Println("Starting proxy server on", *addr)
	if err := http.ListenAndServe(*addr, s); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

type forwardProxy struct{}

func (p *forwardProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodConnect:
		proxyConnect(w, r)
	default:
		http.Error(w, "this proxy only supports CONNECT", http.StatusMethodNotAllowed)
	}
}

func proxyConnect(w http.ResponseWriter, r *http.Request) {
	log.Printf("CONNECT requested to %v (from %v)\n", r.Host, r.RemoteAddr)
	targetConn, err := net.Dial("tcp", r.Host)
	if err != nil {
		log.Printf("failed to dial", r.Host)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)
	hj, ok := w.(http.Hijacker)
	if !ok {
		log.Fatalf("http server does not support hijack %v", err)
	}
	clientConn, _, err := hj.Hijack()
	if err != nil {
		log.Fatal("hijack failed %v", err)
	}
	log.Println("tunnel established")
	go tunnelConnect(targetConn, clientConn)
	go tunnelConnect(clientConn, targetConn)
}

func tunnelConnect(dst io.WriteCloser, src io.ReadCloser) {
	io.Copy(dst, src)
	if err := dst.Close(); err != nil {
		log.Println(err)
	}
	if err := src.Close(); err != nil {
		log.Println(err)
	}
}
