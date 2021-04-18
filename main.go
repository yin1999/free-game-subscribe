package main

import (
	"net/http"
	"os"
)

var (
	accessOrigin string
)

func main() {
	os.Stderr.Write([]byte("starting server...\n"))

	http.HandleFunc("/", handler)
	port := os.Getenv("PORT")
	accessOrigin = os.Getenv("origin")
	if port == "" {
		port = "8080"
	}
	if err := http.ListenAndServe(":"+port, nil); err != nil {
		os.Stderr.Write([]byte(err.Error()))
		os.Exit(-1)
	}
}
