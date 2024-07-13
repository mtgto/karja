package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed index.html
var html string

type ReverseProxyService struct {
	// TODO: Store information about connected other containers
}

func main() {
	mux := http.NewServeMux()
	service := &ReverseProxyService{}
	mux.Handle("/", service)
	log.Fatal(http.ListenAndServe(":9000", mux))
}

func (s *ReverseProxyService) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	html, err := template.New("index").Parse(html)
	if err != nil {
		log.Fatal(err)
	}
	if err := html.Execute(w, nil); err != nil {
		log.Fatal(err)
	}
}
