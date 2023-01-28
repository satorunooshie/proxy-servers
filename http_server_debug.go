package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"strings"
)

func main() {
	if err := debugRequestHeaders(); err != nil {
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
