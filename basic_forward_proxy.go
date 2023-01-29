package main

import (
	"errors"
	"flag"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
)

// Hop-by-hop headers. These are removed when sent to the backend.
// https://www.rfc-editor.org/rfc/rfc9110#name-connection
var hopByHopHeaders = []string{
	"Proxy-Connection",
	"Keep-Alive",
	"TE",
	"Transfer-Encoding",
	"Upgrade",
}

// http_proxy=http://localhost:9999 curl http://example.com
func main() {
	log.SetFlags(log.LUTC | log.Lshortfile)
	s := newForwardProxyServer()
	log.Println("Starting proxy server on", s.Addr)
	if err := http.ListenAndServe(s.Addr, s.Handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func newForwardProxyServer() *http.Server {
	addr := flag.String("addr", "127.0.0.1:9999", "proxy address")
	flag.Parse()
	return &http.Server{
		Addr:    *addr,
		Handler: &forwardProxy{},
	}
}

type forwardProxy struct{}

func (p *forwardProxy) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// The "Host:" header is promoted to Request.Host and is removed from request.Header by net/http, so we print it out explicitly.
	log.Println(r.RemoteAddr, "\t\t", r.Method, "\t\t", r.URL, "\t\t Host:", r.Host)
	log.Println("\t\t\t\t", r.Header)
	if r.URL.Scheme != "http" && r.URL.Scheme != "https" {
		err := errors.New("unsupported protocol scheme " + r.URL.Scheme)
		http.Error(w, err.Error(), http.StatusBadRequest)
		log.Println(err)
		return
	}
	client := http.DefaultClient
	// When an http.Request is sent through an http.Client, RequestURI should not be set.
	// https://www.rfc-editor.org/rfc/rfc9110#section-7.7-4
	r.RequestURI = ""
	if clientIP, _, err := net.SplitHostPort(r.RemoteAddr); err == nil {
		appendHostToXForwardHeader(r.Header, clientIP)
	}

	resp, err := client.Do(r)
	if err != nil {
		http.Error(w, "Server Error", http.StatusInternalServerError)
		log.Fatal("ServeHTTP:", err)
	}
	defer resp.Body.Close()

	log.Println(r.RemoteAddr, " ", resp.Status)

	removeHopHeaders(resp.Header)
	removeConnectionHeaders(resp.Header)
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	io.Copy(w, resp.Body)
}

func copyHeader(dst, src http.Header) {
	for k, vv := range src {
		for _, v := range vv {
			dst.Add(k, v)
		}
	}
}

func removeHopHeaders(header http.Header) {
	for _, h := range hopByHopHeaders {
		header.Del(h)
	}
}

func removeConnectionHeaders(h http.Header) {
	for _, f := range h["Connection"] {
		for _, sf := range strings.Split(f, ",") {
			if sf = strings.TrimSpace(sf); sf != "" {
				h.Del(sf)
			}
		}
	}
}

func appendHostToXForwardHeader(header http.Header, host string) {
	if prior, ok := header["X-Forwarded-For"]; ok {
		host = strings.Join(prior, ",") + ", " + host
	}
	header.Set("X-Forwarded-For", host)
}
