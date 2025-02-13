package main

import (
	"fmt"
	"net/http"
	"os"
	"strings"
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
	protocols = new(http.Protocols)
	protocols.SetHTTP1(true)
	protocols.SetUnencryptedHTTP2(true)
	server := &http.Server{
		Addr:      ":" + port,
		Handler:   handler,
		Protocols: protocols,
	}

	if err := server.ListenAndServe(); err != nil {
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
