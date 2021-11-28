package main

import (
	"io"
	"log"
	"net/http"
)

func main() {

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		io.WriteString(w, "hello")
	})

	err := http.ListenAndServeTLS(":8443", "./cert.pem", "./cert-key.pem", mux)
	if err != nil {
		log.Fatal(err)
	}
}
