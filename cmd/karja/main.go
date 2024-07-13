package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
)

//go:embed index.html
var html string

type ReverseProxyService struct {
	// TODO: Store information about connected other containers
	proxy *httputil.ReverseProxy
}

func main() {
	mux := http.NewServeMux()
	otherContainerUrl, err := url.Parse("http://localhost:3000")
	if err != nil {
		log.Fatal(err)
	}
	proxy := httputil.NewSingleHostReverseProxy(otherContainerUrl)
	service := &ReverseProxyService{proxy}
	mux.Handle("/", service)
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func (s *ReverseProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.Host, "test.") {
		s.proxy.ServeHTTP(w, r)
	} else {
		html, err := template.New("index").Parse(html)
		if err != nil {
			log.Fatal(err)
		}
		if err := html.Execute(w, nil); err != nil {
			log.Fatal(err)
		}
	}
}
