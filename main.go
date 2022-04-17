package main

import (
	"fmt"
	"net/http"
	"os"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var accessOrigin = os.Getenv("origin")

func main() {
	os.Stderr.WriteString("starting server...\n")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	handler := http.HandlerFunc(ServeHTTP)
	h1s := &http.Server{
		Addr:    ":" + port,
		Handler: h2c.NewHandler(handler, &http2.Server{}),
	}

	if err := h1s.ListenAndServe(); err != nil {
		fmt.Fprintln(os.Stderr, err.Error())
		os.Exit(1)
	}
}
