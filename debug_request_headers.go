package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	log.SetFlags(log.LUTC | log.Lshortfile)

	addr := flag.String("addr", "127.0.0.1:7000", "listen address")
	flag.Parse()

	s, err := debugRequestHeaders(*addr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Starting server on", s.Addr)
	if err := http.ListenAndServe(s.Addr, s.Handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

// debugRequestHeaders answers all paths successfully and dumps the request to logs.
// TLS termination with an off-the-shelf reverse proxy.
// ex) caddy reverse-proxy --from :9090 --to :7000.
func debugRequestHeaders(addr string) (*http.Server, error) {
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
	return &http.Server{Addr: addr}, nil
}
