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
	flag.Parse()

	s, err := loadBalancingReverseProxy()
	if err != nil {
		log.Println(err)
		return
	}
	log.Println("Starting proxy server on", s.Addr)
	if err := http.ListenAndServe(s.Addr, s.Handler); err != nil {
		log.Fatal("ListenAndServe:", err)
	}
}

func loadBalancingReverseProxy() (*http.Server, error) {
	toAddr1 := flag.String("to1", "127.0.0.1:7000", "the first address this proxy will forward to")
	toAddr2 := flag.String("to2", "127.0.0.1:7001", "the second address this proxy will forward to")
	fromAddr := flag.String("from", "127.0.0.1:9091", "proxy's listening address")
	flag.Parse()
	t1, err := parseToURL(*toAddr1)
	if err != nil {
		return nil, err
	}
	t2, err := parseToURL(*toAddr2)
	if err != nil {
		return nil, err
	}
	tNum := 1
	return &http.Server{
		Addr: *fromAddr,
		Handler: &httputil.ReverseProxy{
			Director: func(r *http.Request) {
				var target *url.URL
				// Simple round robin between the two targets
				switch tNum {
				case 1:
					target = t1
					tNum = 2
				default:
					target = t2
					tNum = 1
				}
				r.URL.Scheme = target.Scheme
				r.URL.Host = target.Host
				r.URL.Path, r.URL.RawPath = joinURLPath(target, r.URL)
				rawQ := target.RawQuery
				if rawQ == "" || r.URL.RawQuery == "" {
					r.URL.RawQuery = rawQ + r.URL.RawQuery
				} else {
					r.URL.RawQuery = rawQ + "&" + r.URL.RawQuery
				}
				if _, ok := r.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					r.Header.Set("User-Agent", "")
				}
			},
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

// singleJoiningSlash is taken from net/http/httputil.
func singleJoiningSlash(a, b string) string {
	aslash := strings.HasSuffix(a, "/")
	bslash := strings.HasPrefix(b, "/")
	switch {
	case aslash && bslash:
		return a + b[1:]
	case !aslash && !bslash:
		return a + "/" + b
	}
	return a + b
}

// joinURLPath is taken from net/http/httputil.
func joinURLPath(a, b *url.URL) (path, rawpath string) {
	if a.RawPath == "" && b.RawPath == "" {
		return singleJoiningSlash(a.Path, b.Path), ""
	}
	// Same as singleJoiningSlash, but uses EscapedPath to determine
	// whether a slash should be added
	apath := a.EscapedPath()
	bpath := b.EscapedPath()

	aslash := strings.HasSuffix(apath, "/")
	bslash := strings.HasPrefix(bpath, "/")

	switch {
	case aslash && bslash:
		return a.Path + b.Path[1:], apath + bpath[1:]
	case !aslash && !bslash:
		return a.Path + "/" + b.Path, apath + "/" + bpath
	}
	return a.Path + b.Path, apath + bpath
}
