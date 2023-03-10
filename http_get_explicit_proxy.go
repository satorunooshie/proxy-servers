package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
)

// Go HTTP clients will not proxy localhost addresses (or their synonyms) by default, even if http_proxy is set.
// To avoid triggering the case above, Go's http.Client can be configured with a http.transport,
// which by default uses ProxyFromEnvironment for a custom proxy configuration,
// without mucking with the machine's network configuration(127.0.0.1 local.alias >> /etc/hosts).
// go run debug_request_headers.go --addr localhost:8080
// go run basic_forward_proxy.go
// go run http_get_explicit_proxy.go --target http://localhost:8080/foo/bar
func main() {
	log.SetFlags(log.LUTC | log.Lshortfile)

	proxy := flag.String("proxy", "http://localhost:9999", "proxy to use")
	target := flag.String("target", "http://example.org", "URL to get")
	flag.Parse()

	proxyURL, err := url.Parse(*proxy)
	if err != nil {
		log.Println(err)
		return
	}
	client := &http.Client{
		Transport: &http.Transport{Proxy: http.ProxyURL(proxyURL)},
	}
	resp, err := client.Get(*target)
	if err != nil {
		log.Println(err)
		return
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Println(err)
		return
	}
	fmt.Printf("%s\n", body)
}
