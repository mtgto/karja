package main

import (
	"log"
	"net/http"
)

func main() {
	http.HandleFunc("/", baseHandler)
	log.Fatal(http.ListenAndServe(":8080", nil))
}

func baseHandler(w http.ResponseWriter, r *http.Request) {
	hello := []byte("Hello, world!")
	_, err := w.Write(hello)
	if err != nil {
		log.Fatal(err)
	}
}
