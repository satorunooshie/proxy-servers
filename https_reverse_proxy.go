package main

import (
	"crypto/tls"
	"flag"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

// mkcert -install local.alias
// go run https_reverse_proxy.go -cert local.alias.pem -key local.alias-key.pem
// go run debug_request_headers.go
// curl -Lv https://local.alias:9000/hogehoge
func main() {
	log.SetFlags(log.LUTC | log.Lshortfile)

	fromAddr := flag.String("from", "127.0.0.1:9090", "proxy's listening address")
	toAddr := flag.String("to", "127.0.0.1:7000", "the address this proxy will forward to")
	cert := flag.String("cert", "cert.pem", "certificate PEM file")
	key := flag.String("key", "key.pem", "key PEM file")
	flag.Parse()

	s, err := httpsReverseProxy(*fromAddr, *toAddr)
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Starting proxy server on", s.Addr)
	if err := s.ListenAndServeTLS(*cert, *key); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func httpsReverseProxy(from, to string) (*http.Server, error) {
	toURL, err := parseToURL(to)
	if err != nil {
		return nil, err
	}
	return &http.Server{
		Addr:    from,
		Handler: httputil.NewSingleHostReverseProxy(toURL),
		TLSConfig: &tls.Config{
			MinVersion:               tls.VersionTLS13,
			PreferServerCipherSuites: true,
		},
	}, nil
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
