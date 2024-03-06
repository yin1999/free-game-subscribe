package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

var accessOrigins = map[string]struct{}{}

func main() {
	os.Stderr.WriteString("starting server...\n")
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	resolveAccessOrigins()

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

func resolveAccessOrigins() {
	env := os.Getenv("origin")
	if env == "" {
		return
	}
	for _, origin := range strings.Split(env, ",") {
		origin = strings.TrimSpace(origin)
		accessOrigins[origin] = struct{}{}
	}
}
