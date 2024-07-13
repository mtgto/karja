package main

import (
	_ "embed"
	"html/template"
	"log"
	"net/http"
)

//go:embed index.html
var html string

func main() {
	http.HandleFunc("/", baseHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	html, err := template.New("index").Parse(html)
	if err != nil {
		log.Fatal(err)
	}
	if err := html.Execute(w, nil); err != nil {
		log.Fatal(err)
	}
}
