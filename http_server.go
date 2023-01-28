package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

func main() {
	if err := basicReverseProxy(); err != nil {
		log.Fatal(err)
	}
}

// debugRequestHeaders answers all paths successfully and dumps the request to logs.
// TLS termination with an off-the-shelf reverse proxy.
// ex) caddy reverse-proxy --from :9090 --to :7000.
func debugRequestHeaders() error {
	addr := flag.String("addr", "127.0.0.1:7000", "listen address")
	flag.Parse()
	http.HandleFunc("/",
		func(w http.ResponseWriter, r *http.Request) {
			var b strings.Builder
			_, _ = fmt.Fprintf(&b, "%v\t%v\t%v\tHost:%v\n", r.RemoteAddr, r.Method, r.URL, r.Host)
			for name, headers := range r.Header {
				for _, h := range headers {
					_, _ = fmt.Fprintf(&b, "%v: %v\n", name, h)
				}
			}
			log.Println(b.String())
			fmt.Fprintf(w, "Hello %s\n", r.URL)
		},
	)
	log.Println("Starting server on", *addr)
	if err := http.ListenAndServe(*addr, nil); err != nil {
		return fmt.Errorf("ListenAndServe: %v", err)
	}
	return nil
}

func basicReverseProxy() error {
	fromAddr := flag.String("from", "localhost:9090", "proxy's listening address")
	toAddr := flag.String("to", "localhost:7000", "the address this proxy will forward to")
	flag.Parse()
	toURL, err := parseToURL(*toAddr)
	if err != nil {
		return err
	}
	proxy := httputil.NewSingleHostReverseProxy(toURL)
	log.Println("Starting proxy server on", *fromAddr)
	if err := http.ListenAndServe(*fromAddr, proxy); err != nil {
		return fmt.Errorf("ListenAndServe: %v", err)
	}
	return nil
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
